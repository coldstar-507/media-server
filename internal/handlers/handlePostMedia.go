package handlers

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	// "image"
	// "image/png"

	// "image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coldstar-507/flatgen"
	"github.com/coldstar-507/media-server/internal/config"
	"github.com/coldstar-507/media-server/internal/paths"
	media_utils "github.com/coldstar-507/media-server/internal/utils"
	"github.com/coldstar-507/utils2"
	// "github.com/coldstar-507/utils/utils"
	"golang.org/x/sync/errgroup"
	// "golang.org/x/sync/errgroup"
	// "github.com/gographics/imagick/imagick"
	// "github.com/h2non/bimg"
	// "github.com/disintegration/imaging"
)

type MediaWriteRequest struct {
	metadata []byte
	data     io.Reader
	perm     bool
	thumsize int
	res      chan error
}

var mwr = make(chan *MediaWriteRequest)

type ErrorCollector struct {
	mu   sync.Mutex
	errs []error
}

func (ec *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errs = append(ec.errs, err)
}

func (ec *ErrorCollector) Err() error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return errors.Join(ec.errs...)
}

func RunMediaWriteRequestsHandler() {
	log.Println("Running MediaWriteRequestsHandler")
	tmp, err := os.OpenFile(paths.TEMP_IDS_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	bb := make([]byte, 0, 80)
	idbuf := bytes.NewBuffer(bb)
	var idf *os.File

	daily := time.NewTicker(time.Hour * 24)

	for {
		select {
		case req := <-mwr:
			log.Printf("MediaWriteRequest")
			idf = utils2.If(req.perm, nil, tmp)
			req.res <- handleWriteRequest(req, idbuf, idf)

		case <-daily.C:
			log.Printf("Daily delete routine")
			nDelete := deleteRoutine(tmp)
			tmp.Close()
			if err := deleteFirstNLines(paths.TEMP_IDS_FILE, nDelete); err != nil {
				log.Println("deleteFirstNLines err:", err)
			}
			tmp, err = os.OpenFile(paths.TEMP_IDS_FILE,
				os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Println("open tmp file error:", err)
			}
		}
	}
}

func deleteFirstNLines(filePath string, n int) error {
	// Open the original file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var remainingLines []string
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		txt := scanner.Text()
		if lineCount >= n && len(txt) > 0 {
			remainingLines = append(remainingLines, txt)
		}
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Open the same file for writing (truncate it first)
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range remainingLines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func handleWriteRequest(req *MediaWriteRequest, idbuf *bytes.Buffer, idf *os.File) error {
	mr := flatgen.GetRootAsMediaMetadata(req.metadata, 0)
	ref := mr.Reference(nil)
	ref.MutateTimestamp(time.Now().UnixMilli())

	utils2.WriteMediaReference(idbuf, ref)
	defer idbuf.Reset()
	strid := hex.EncodeToString(idbuf.Bytes())
	ogExt := media_utils.OriginalExtOf(strid)
	// svExt := media_utils.Serv
	if idf != nil {
		_, err := idf.WriteString(strid + "\n")
		if err != nil {
			return fmt.Errorf("write id=%s error: %w", strid, err)
		}
	}

	metadatafile, err := os.Create(paths.MakePath("meta", strid))
	if err != nil {
		return fmt.Errorf("metadata create file error: %w", err)
	}
	defer metadatafile.Close()
	if _, err := metadatafile.Write(req.metadata); err != nil {
		return fmt.Errorf("metadata file write error: %w", err)
	}

	if mr.HasData() {
		tempp := paths.MakePathExt("temp", strid, ogExt)
		tempf, err := os.Create(tempp)
		if err != nil {
			return fmt.Errorf("create files error: %w", err)
		}
		defer tempf.Close()
		defer os.Remove(tempp)
		if _, err := io.Copy(tempf, req.data); err != nil {
			return fmt.Errorf("copying to temp file error: %w", err)
		}

		g := new(errgroup.Group)
		g.Go(func() error {
			mainp := paths.MakePathExt("data", strid, "webp")
			return media_utils.ToWebp(tempp, mainp)
		})
		if size := req.thumsize; size > 0 {
			g.Go(func() error {
				thump := paths.MakePathExt("thum", strid, "webp")
				return media_utils.ToWebpSquareThumbnail(tempp, thump, size)
			})
		}
		return g.Wait()

		// g.Go(func() error {
		// 	mainavif := paths.MakePathExt("data", strid, svExt)
		// 	return media_utils.ToAvif(temp, mainavif)

		// })
		// if req.thumsize > 0 {
		// 	g.Go(func() error {
		// 		thump := paths.MakePathExt("thumbnail", strid, svExt)
		// 		return media_utils.ToAvifSquareThumbnail(temp, thump)
		// 	})
		// }
		// return g.Wait()
	}
	return nil
}

func deleteRoutine(idf *os.File) int {
	buf := make([]byte, 0, 8)
	timeThreshold := time.Now().Add(-time.Hour * 24 * 14).UnixMilli()
	buf, _ = binary.Append(buf, binary.BigEndian, timeThreshold)
	twoWeeksAgo := hex.EncodeToString(buf)
	scanner := bufio.NewScanner(idf)
	var ix int
	var toDelete []string
	for scanner.Scan() {
		id := scanner.Text()
		ts := id[2 : 2+2*8]
		doDelete := ts < twoWeeksAgo
		if !doDelete {
			break
		}
		toDelete = append(toDelete, id)
		ix++
	}

	go func() {
		for _, id := range toDelete {
			DeleteMedia(id)
		}
	}()
	return ix
}

func DeleteMedia(id string) {
	log.Printf("DeleteMedia(%s)\n", id)
	thum := paths.MakePathExt("thum", id, media_utils.ServerExtOfId(id))
	data := paths.MakePathExt("data", id, media_utils.ServerExtOfId(id))
	meta := paths.MakePath("meta", id)
	os.Remove(thum)
	os.Remove(data)
	os.Remove(meta)
}

func HandlePostMedia(w http.ResponseWriter, r *http.Request) {
	thum, _ := strconv.Atoi(r.PathValue("thum"))
	var mdlen uint16
	err0 := utils2.ReadBin(r.Body, &mdlen)
	metadata := make([]byte, mdlen)
	err1 := utils2.ReadBin(r.Body, metadata)
	if err := errors.Join(err0, err1); err != nil {
		w.WriteHeader(500)
		return
	}
	meta := flatgen.GetRootAsMediaMetadata(metadata, 0)
	ref := meta.Reference(nil)
	ref.MutatePlace(config.Config.SERVER_PLACE)

	log.Printf("HandlePostMedia3, hasData=%v, thum=%d, perm=%v\n",
		meta.HasData(), thum, ref.Perm())

	ch := make(chan error)
	defer close(ch)
	mwr <- &MediaWriteRequest{
		res:      ch,
		perm:     ref.Perm(),
		metadata: metadata,
		data:     utils2.If(meta.HasData(), r.Body, nil),
		thumsize: thum,
	}
	if err := <-ch; err != nil {
		log.Println("HandlePostMedia2: error:", err)
		w.WriteHeader(501)
	}
	utils2.WriteMediaReference(w, ref)
}

// type MediaWriteRequest struct {
// 	fullMedia *flatgen.FullMedia
// 	perm      bool
// 	res       chan error
// }
// type MediaWriteRequest struct {
// 	format        imaging.Format
// 	metadata      []byte
// 	data          io.Reader
// 	perm          bool
// 	w, h, t       int
// 	makeThumbnail bool
// 	res           chan error
// }

// type MediaWriteRequest2 struct {
// 	// format        imaging.Format
// 	metadata, data []byte
// 	perm           bool
// 	thumsize       int
// 	res            chan error
// }

// var mwr2 = make(chan *MediaWriteRequest2)

// func transformImage(path string, data io.Reader, w, h, t int) error {
// 	log.Printf("transformImage path=%s, w=%d, h=%d, t=%d\n", path, w, h, t)
// 	if t > 2 {
// 		return fmt.Errorf("invalid transform type=%d", t)
// 	}
// 	file, err := os.Create(path)
// 	if err != nil {
// 		return fmt.Errorf("create thumbnail file error: %w", err)
// 	}

// 	im, err := imaging.Decode(data)
// 	if err != nil {
// 		return fmt.Errorf("decode image error: %w", err)
// 	}

// 	var rz *image.NRGBA
// 	switch t {
// 	case 0:
// 		rz = imaging.Fill(im, w, h, imaging.Center, imaging.Lanczos)
// 	case 1:
// 		rz = imaging.Fit(im, w, h, imaging.Lanczos)
// 	case 2:
// 		rz = imaging.Resize(im, w, h, imaging.Lanczos)
// 	}
// 	if err = imaging.Encode(file, rz, imaging.JPEG); err != nil {
// 		return fmt.Errorf("encode image error: %w", err)
// 	}
// 	return file.Close()
// }

// func handleWriteRequest(req *MediaWriteRequest, idbuf *bytes.Buffer, idf *os.File) error {
// 	mr := flatgen.GetRootAsMediaMetadata(req.metadata, 0)
// 	ref := mr.Reference(nil)
// 	ref.MutateTimestamp(time.Now().UnixMilli())

// 	utils2.WriteMediaReference(idbuf, ref)
// 	defer idbuf.Reset()
// 	strid := hex.EncodeToString(idbuf.Bytes())
// 	if idf != nil {
// 		_, err := idf.WriteString(strid + "\n")
// 		if err != nil {
// 			return fmt.Errorf("write id=%s error: %w", strid, err)
// 		}
// 	}

// 	metadatafile, err := os.Create(paths.MakePath("meta", strid))
// 	if err != nil {
// 		return fmt.Errorf("metadata create file error: %w", err)
// 	}
// 	defer metadatafile.Close()
// 	if _, err := metadatafile.Write(req.metadata); err != nil {
// 		return fmt.Errorf("metadata file write error: %w", err)
// 	}

// 	if mr.HasData() {
// 		mediap := paths.MakePath("data", strid)
// 		datafile, err := os.Create(mediap)
// 		if err != nil {
// 			return fmt.Errorf("create files error: %w", err)
// 		}
// 		defer datafile.Close()
// 		im, err := imaging.Decode(req.data)
// 		if err != nil {
// 			return fmt.Errorf("main image decode error: %w", err)
// 		}

// 		x, y := im.Bounds().Dx(), im.Bounds().Dy()
// 		mn, mx := min(x, y), max(x, y)
// 		// if static image, png, jpeg, ...
// 		if mn <= 256 && mx <= 4*256 {
// 			// small or pixel art -> png nearest
// 			err = imaging.Encode(datafile, im, imaging.PNG)
// 		} else if mn <= 512 && mx <= 3*512 {
// 			// medium, bicubic webp
// 			err = imaging.Encode(datafile, im, imaging.PNG,
// 				imaging.PNGCompressionLevel(png.DefaultCompression))
// 		} else {
// 			var rz image.Image
// 			if x > y {
// 				rz = imaging.Resize(im, maxsize, 0, imaging.Lanczos)
// 			} else {
// 				rz = imaging.Resize(im, 0, maxsize, imaging.Lanczos)
// 			}
// 			err = imaging.Encode(datafile, rz, imaging.PNG,
// 				imaging.PNGCompressionLevel(png.BestCompression))
// 		}

// 		if err != nil {
// 			return fmt.Errorf("main image encode error: %w", err)
// 		}
// 		if req.makeThumbnail {
// 			tb := paths.MakePath("thum", strid)
// 			datafile.Seek(0, io.SeekStart)
// 			if err := transformImage(tb, datafile, req.w, req.h, req.t); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func WriteMedia(id string, permanent bool, rdr io.Reader) error {
// 	path := paths.MakeMediaPath(id, permanent)
// 	f, err := os.Create(path)
// 	if err != nil {
// 		return fmt.Errorf("WriteMedia error creating media file id=%s, %v", id, err)
// 	}
// 	defer f.Close()
// 	if _, err := io.Copy(f, rdr); err != nil && err != io.EOF {
// 		return fmt.Errorf("WriteMedia error saving media file id=%s: %v", id, err)
// 	}
// 	return nil
// }

// func HandlePostMedia(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")
// 	permanent := paths.IsPermanent(id)
// 	defer r.Body.Close()
// 	if err := WriteMedia(id, permanent, r.Body); err != nil {
// 		log.Println("HandlePostMedia error: ", err)
// 		w.WriteHeader(500)
// 	}
// }

// func HandlePostMedia2(w http.ResponseWriter, r *http.Request) {
// 	ext := r.PathValue("ext")
// 	format, err := imaging.FormatFromExtension(ext)
// 	width, _ := strconv.Atoi(r.PathValue("width"))
// 	height, _ := strconv.Atoi(r.PathValue("height"))
// 	t, _ := strconv.Atoi(r.PathValue("type"))

// 	var mdlen uint16
// 	err0 := utils2.ReadBin(r.Body, &mdlen)
// 	metadata := make([]byte, mdlen)
// 	err1 := utils2.ReadBin(r.Body, metadata)
// 	if err := errors.Join(err0, err1); err != nil {
// 		w.WriteHeader(500)
// 		return
// 	}
// 	meta := flatgen.GetRootAsMediaMetadata(metadata, 0)
// 	ref := meta.Reference(nil)
// 	makeThumbnail := format >= 0 && (width > 0 || height > 0)

// 	log.Printf("HandlePostMedia2, w=%d, h=%d, t=%d, ext=%s, fmt=%d"+
// 		" hasData=%v, perm=%v, makeThumb=%v\n",
// 		width, height, t, ext, format, meta.HasData(), ref.Perm(), makeThumbnail)

// 	ch := make(chan error)
// 	defer close(ch)
// 	mwr <- &MediaWriteRequest{
// 		format:        format,
// 		res:           ch,
// 		perm:          ref.Perm(),
// 		metadata:      metadata,
// 		data:          utils2.If(meta.HasData(), r.Body, nil),
// 		w:             width,
// 		h:             height,
// 		t:             t,
// 		makeThumbnail: format >= 0 && (width > 0 || height > 0),
// 	}
// 	if err = <-ch; err != nil {
// 		log.Println("HandlePostMedia2: error:", err)
// 		w.WriteHeader(501)
// 	}
// 	utils2.WriteMediaReference(w, ref)
// }

// w, h := meta.Size.Width, meta.Size.Height

// const maxSize uint = 1296

// func handleWriteRequest2(req *MediaWriteRequest2, idbuf *bytes.Buffer, idf *os.File) error {
// 	mr := flatgen.GetRootAsMediaMetadata(req.metadata, 0)
// 	ref := mr.Reference(nil)
// 	ref.MutateTimestamp(time.Now().UnixMilli())

// 	utils2.WriteMediaReference(idbuf, ref)
// 	defer idbuf.Reset()
// 	strid := hex.EncodeToString(idbuf.Bytes())
// 	if idf != nil {
// 		_, err := idf.WriteString(strid + "\n")
// 		if err != nil {
// 			return fmt.Errorf("write id=%s error: %w", strid, err)
// 		}
// 	}

// 	metadatafile, err := os.Create(paths.MakePath("meta", strid))
// 	if err != nil {
// 		return fmt.Errorf("metadata create file error: %w", err)
// 	}
// 	defer metadatafile.Close()
// 	if _, err := metadatafile.Write(req.metadata); err != nil {
// 		return fmt.Errorf("metadata file write error: %w", err)
// 	}

// 	if mr.HasData() {

// 		format, animated, width, height, err := media_utils.InspectImage(req.data)

// 		// var resize bool = false
// 		w, h := width, height

// 		datapath := paths.MakePath("data", strid)
// 		datafile, err := os.Create(datapath)
// 		if err != nil {
// 			return fmt.Errorf("create datafile error: %w", err)
// 		}
// 		defer datafile.Close()

// 		if !animated {

// 			if m := max(width, height); m > int(maxSize) {
// 				// resize = true
// 				if m == w {
// 					w = int(maxSize)
// 					h = 0
// 				} else {
// 					h = int(maxSize)
// 					w = 0
// 				}
// 			}

// 			bm := bimg.NewImage(req.data)
// 			opts := bimg.Options{
// 				Quality:     85,
// 				Width:       w,
// 				Height:      h,
// 				Compression: 9,
// 			}

// 			t, err := bm.Process(opts)
// 			if err != nil {
// 				return fmt.Errorf("process image error: %w", err)
// 			} else if t == nil {
// 				return errors.New("empty result from image processing")
// 			}

// 			_, err = datafile.Write(t)
// 			if err != nil {
// 				return fmt.Errorf("write data file error: %w", err)
// 			}

// 			if req.thumsize > 0 {
// 				thumpath := paths.MakePath("thum", strid)
// 				thumfile, err := os.Create(thumpath)
// 				if err != nil {
// 					return fmt.Errorf("create thumfile error: %w", err)
// 				}
// 				defer thumfile.Close()
// 				tnopts := bimg.Options{
// 					Quality:     85,
// 					Width:       req.thumsize,
// 					Height:      req.thumsize,
// 					Compression: 9,
// 					Crop:        true,
// 					Gravity:     bimg.GravityCentre,
// 				}
// 				tn, err := bm.Process(tnopts)
// 				if err != nil {
// 					return fmt.Errorf("tn process error: %w", err)
// 				} else if tn == nil {
// 					return errors.New("empty result from tn processing")
// 				}

// 				_, err = thumfile.Write(t)
// 				if err != nil {
// 					return fmt.Errorf("write thum file error: %w", err)
// 				}
// 			}
// 		} else {
// 			var err error
// 			wand := imagick.NewMagickWand()
// 			defer wand.Destroy()
// 			err = wand.ReadImageBlob(req.data)
// 			if err != nil {
// 				return fmt.Errorf("error reading blob: %w", err)
// 			}

// 			if uint(max(w, h)) > maxSize {
// 				wand, err = media_utils.ResizeAnimated(wand, maxSize)
// 				if err != nil {
// 					return fmt.Errorf("error resizing to maxSize: %w", err)
// 				}
// 			}

// 			wand, err = media_utils.CompressAnimated(wand, format)
// 			if err != nil {
// 				return fmt.Errorf("error compressing format=%s: %w", format, err)
// 			}

// 			err = wand.WriteImagesFile(datafile)
// 			if err != nil {
// 				return fmt.Errorf("error writingImagesFile to datafile: %w", err)
// 			}

// 			if req.thumsize > 0 {
// 				thumpath := paths.MakePath("thum", strid)
// 				thumfile, err := os.Create(thumpath)
// 				if err != nil {
// 					return fmt.Errorf("create thumfile error: %w", err)
// 				}
// 				defer thumfile.Close()
// 				wand, err = media_utils.ResizeAnimated(wand, uint(req.thumsize))
// 				if err != nil {
// 					return fmt.Errorf("thum resize error: %w", err)
// 				}
// 				wand, err = media_utils.SquareCropAnimated(wand)
// 				if err != nil {
// 					return fmt.Errorf("thum crop error: %w", err)
// 				}
// 				err = wand.WriteImagesFile(thumfile)
// 				if err != nil {
// 					return fmt.Errorf("thum write images error: %w", err)
// 				}

// 			}

// 		}
// 	}
// 	return nil
// }

// func HandlePostMedia2(w http.ResponseWriter, r *http.Request) {
// 	if buf, err := io.ReadAll(r.Body); err != nil {
// 		w.WriteHeader(502)
// 	} else {
// 		fm := flatgen.GetRootAsFullMedia(buf, 0)
// 		ref := flatgen.GetRootAsMediaMetadata(fm.MetadataBytes(), 0).Ref(nil)
// 		ch := make(chan error)
// 		defer close(ch)
// 		mwr <- &MediaWriteRequest{
// 			fullMedia: fm,
// 			res:       ch,
// 			perm:      ref.Permanent(),
// 		}
// 		if err = <-ch; err != nil {
// 			w.WriteHeader(503)
// 		}
// 		utils2.WriteMediaRef(w, ref)
// 	}
// }
