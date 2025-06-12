package media_utils

import (
	// "bytes"
	// "encoding/json"
	// "bytes"
	"fmt"
	"io"
	"os/exec"
	// "strings"
	// "github.com/coldstar-507/utils/utils"
)

const maxside int = 1280

var strExtMap = map[string]string{
	"0001": "jpg",
	"0002": "png",
	"0003": "gif",
	"0004": "webp",
	"0005": "tif",
	"0006": "bmp",
	"0007": "avif",
}

var conversionMap = map[string]string{
	"jpg":  "webp",
	"png":  "webp",
	"gif":  "webp",
	"webp": "webp",
	"tif":  "webp",
	"bmp":  "webp",
	"avif": "webp",
}

func OriginalExtOf(id string) string {
	// id is a media ref
	// it has, 1 byte for permanence
	// 2 byte for place
	// in hex, so mult by 2
	// type is 2 bytes
	l := len(id)
	k := id[l-10 : l-6] // this should be the ext
	ext := strExtMap[k]
	fmt.Printf("OriginalExtOf(%s): %s\n", id, ext)
	return ext
}

func ServerExtOfId(id string) string {
	og := OriginalExtOf(id)
	ext := conversionMap[og]
	fmt.Printf("ServerExtOfId(%s): %s\n", id, ext)
	return ext
}

func ToAvifSquareThumbnail(inpath, outpath string) error {
	const cropfilter = `crop='min(iw\,ih)':min(iw\,ih)',scale=256:256`
	cmd := exec.Command("ffmpeg",
		"-i", inpath,
		"-vf", cropfilter,
		"-an",
		"-c:v", "libaom-av1",
		"-crf", "35",
		"-b:v", "0",
		"-pix_fmt", "yuv420p",
		"-map_metadata", "-1",
		"-movflags", "+faststart",
		outpath,
	)

	fmt.Printf("ToAvifSquareThumbnail: running: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("thumbnail error: %w, output=%s\n", err, string(out))
	}
	return nil
}

func ToAvif(inpath, outpath string) error {
	filter := `scale=iw*min(1\,min(1280/iw\,1280/ih)):ih*min(1\,min(1280/iw\,1280/ih))`
	// filter := "scale=w=1280:h=1280:force_original_aspect_ratio=decrease"
	cmd := exec.Command("ffmpeg",
		"-i", inpath,
		"-vf", filter,
		"-an",
		"-c:v", "libaom-av1",
		"-crf", "35",
		"-b:v", "0",
		"-pix_fmt", "yuv420p",
		"-map_metadata", "-1",
		"-movflags", "+faststart",
		outpath,
	)

	fmt.Printf("ToAvif: running: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run error: %w, output=%s\n", err, string(out))
	}
	return nil
}

func PipeToAvif(r io.Reader, outpath string) error {
	const filter = `scale=iw*min(1\,min(1280/iw\,1280/ih)):ih*min(1\,min(1280/iw\,1280/ih))`
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-vf", filter,
		"-an",
		"-c:v", "libaom-av1",
		"-crf", "35",
		"-b:v", "0",
		"-pix_fmt", "yuv420p",
		"-map_metadata", "-1",
		"-movflags", "+faststart",
		outpath,
	)
	fmt.Printf("ToAvifPipe: running: %s\n", cmd.String())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("get stdin error: %w", err)
	}
	defer stdin.Close()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start error: %w", err)
	}
	if _, err := io.Copy(stdin, r); err != nil {
		return fmt.Errorf("copy to stdin error: %w", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("stdin close error: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg error: %w", err)
	}
	return nil
}

func PipeToWebp(r io.Reader, outpath string) error {
	const filter = "scale=iw*min(1\\,min(1280/iw\\,1280/ih)):ih*min(1\\,min(1280/iw\\,1280/ih))"
	// ffmpeg -i transparent.gif -vcodec webp -loop 0 -pix_fmt yuva420p transparent.webp
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-vf", filter,
		"-vcodec", "webp",
		"-loop", "0",
		"-pix_fmt", "yuva420p",
		outpath,
	)
	fmt.Printf("PipeToWebp: running: %s\n", cmd.String())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("get stdin error: %w", err)
	}
	defer stdin.Close()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start error: %w", err)
	}
	if _, err := io.Copy(stdin, r); err != nil {
		return fmt.Errorf("copy to stdin error: %w", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("stdin close error: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg error: %w", err)
	}
	return nil
}

func ToWebp(inpath, outpath string) error {
	const filter = "scale=iw*min(1\\,min(1280/iw\\,1280/ih)):ih*min(1\\,min(1280/iw\\,1280/ih))"
	cmd := exec.Command("ffmpeg",
		"-i", inpath,
		"-vf", filter,
		"-fps_mode", "vfr",
		"-c:v", "libwebp",
		"-lossless", "0",
		"-q:v", "75",
		"-preset", "default",
		"-loop", "0",
		"-pix_fmt", "yuva420p",
		outpath)
	fmt.Printf("ToWebp: running: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run error: %w, output=%s\n", err, string(out))
	}
	return nil
}

func ToWebpSquareThumbnail(inpath, outpath string, size int) error {
	filter := fmt.Sprintf("crop='min(in_w\\,in_h)':'min(in_w\\,in_h)',scale=%d:%d", size, size)
	cmd := exec.Command("ffmpeg",
		"-i", inpath,
		"-vf", filter,
		"-fps_mode", "vfr",
		"-c:v", "libwebp",
		"-lossless", "0",
		"-q:v", "75",
		"-preset", "default",
		"-loop", "0",
		"-pix_fmt", "yuva420p",
		outpath)
	fmt.Printf("ToWebpSquareThumbnail: running: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("thumbnail error: %w, output=%s\n", err, string(out))
	}
	return nil
}

// func ScaleCompress(inputPath, outputPath, ext string, inputw, inputh int) error {
// 	cmdArgs := []string{"-i", inputPath}

// 	var scale string
// 	if max(inputw, inputh) > maxside {
// 		if inputw > inputh {
// 			scale = fmt.Sprintf("scale=%d:-1", maxside)
// 		} else {
// 			scale = fmt.Sprintf("scale=-1:%d", maxside)
// 		}
// 		cmdArgs = append(cmdArgs, "-vf", scale)
// 	}

// 	// Determine format and apply compression flags
// 	switch ext {
// 	case "jpg", "webp":
// 		cmdArgs = append(cmdArgs, "-q:v", fmt.Sprintf("%d", 85))
// 	case "png":
// 		cmdArgs = append(cmdArgs, "-compression_level", fmt.Sprintf("%d", 9))
// 	case "gif", "apng":
// 		cmdArgs = append(cmdArgs, "-loop", "0")
// 	}

// 	cmdArgs = append(cmdArgs, outputPath)
// 	fmt.Printf("ScaleCompress: running: ffmpeg %s\n", strings.Join(cmdArgs, " "))
// 	cmd := exec.Command("ffmpeg", cmdArgs...)
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("run error: %w, output=%s\n", err, string(out))
// 	}
// 	return nil
// }

// type ffprobeOutput struct {
// 	Streams []struct {
// 		Width  int `json:"width"`
// 		Height int `json:"height"`
// 	} `json:"streams"`
// }

// func GetImageDimensions(path string) (width, height int, err error) {
// 	cmd := exec.Command("ffprobe",
// 		"-v", "error",
// 		"-select_streams", "v:0",
// 		"-show_entries", "stream=width,height",
// 		"-of", "json",
// 		path,
// 	)

// 	var out bytes.Buffer
// 	cmd.Stdout = &out

// 	if err := cmd.Run(); err != nil {
// 		return 0, 0, fmt.Errorf("ffprobe failed: %w", err)
// 	}

// 	var result ffprobeOutput
// 	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
// 		return 0, 0, fmt.Errorf("failed to parse ffprobe output: %w", err)
// 	}

// 	if len(result.Streams) == 0 {
// 		return 0, 0, fmt.Errorf("no streams found in image")
// 	}

// 	return result.Streams[0].Width, result.Streams[0].Height, nil
// }

// func MakeThumbnail(inputPath, outputPath, ext string, thumscale, inputw, inputh int) error {
// 	var scale = `crop='min(in_w\,in_h)':'min(in_w\,in_h)'`
// 	if max(inputw, inputh) > maxside {
// 		if inputw > inputh {
// 			scale += fmt.Sprintf(",scale=%d:-1", thumscale)
// 		} else {
// 			scale += fmt.Sprintf(",scale=-1:%d", thumscale)
// 		}
// 	}

// 	cmdArgs := []string{
// 		"-i", inputPath,
// 		"-vf", scale,
// 	}

// 	// Determine format and apply compression flags
// 	switch ext {
// 	case "jpg", "webp":
// 		cmdArgs = append(cmdArgs, "-q:v", fmt.Sprintf("%d", 85))
// 	case "png":
// 		cmdArgs = append(cmdArgs, "-compression_level", fmt.Sprintf("%d", 9))
// 	case "gif", "apng":
// 		cmdArgs = append(cmdArgs, "-loop", "0")
// 	}

// 	cmdArgs = append(cmdArgs, outputPath)
// 	cmd := exec.Command("ffmpeg", cmdArgs...)
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("run error: %w, output=%s\n", err, string(out))
// 	}
// 	return nil
// }

// import (
// 	"bytes"
// 	"fmt"
// 	_ "image/gif"
// 	_ "image/jpeg"
// 	_ "image/png"

// 	"strings"

// 	"github.com/gographics/imagick/imagick"
// 	"github.com/h2non/bimg"
// )

// func InspectImage(data []byte) (format string, animated bool, width, height int, err error) {
// 	buf := bimg.NewImage(data)
// 	meta, err := buf.Metadata()
// 	if err != nil {
// 		return "", false, 0, 0, err
// 	}

// 	format = strings.ToLower(meta.Type)

// 	// libvips sometimes misses animation flag, fallback for GIF
// 	switch format {
// 	case "gif":
// 		animated = bytes.Count(data, []byte("Graphic Control Extension")) > 1
// 	case "webp":
// 		animated = bytes.Contains(data, []byte("ANIM"))
// 	case "png":
// 		animated = bytes.Contains(data, []byte("acTL")) // APNG marker
// 	default:
// 		animated = false
// 	}

// 	return format, animated, meta.Size.Width, meta.Size.Height, nil
// }

// func CompressAnimated(input *imagick.MagickWand, format string) (*imagick.MagickWand, error) {
// 	defer input.Destroy()
// 	coalesced := input.Clone().CoalesceImages()
// 	defer coalesced.Destroy()

// 	compressed := imagick.NewMagickWand()

// 	for i := 0; i < int(coalesced.GetNumberImages()); i++ {
// 		coalesced.SetIteratorIndex(i)

// 		frame := coalesced.GetImage()

// 		frame.SetImageDelay(coalesced.GetImageDelay())
// 		frame.SetImageDispose(coalesced.GetImageDispose())

// 		switch format {
// 		case "gif":
// 			frame.SetImageFormat("gif")
// 			frame.SetImageCompression(imagick.COMPRESSION_LZW)
// 			frame.SetImageCompressionQuality(75)
// 			break
// 		case "webp":
// 			frame.SetImageFormat("webp")
// 			frame.SetImageCompressionQuality(75)
// 			break
// 		case "png", "apgn":
// 			frame.SetImageFormat("png")
// 			frame.SetImageCompression(imagick.COMPRESSION_ZIP)
// 			frame.SetImageCompressionQuality(8)
// 		}

// 		compressed.AddImage(frame)
// 		frame.Destroy()

// 	}

// 	// Optimize layers to reduce size (optional)
// 	if err := compressed.OptimizeImageLayers(); err != nil {
// 		return nil, fmt.Errorf("failed to optimize layers: %w", err)
// 	}

// 	return compressed, nil

// }

// func SquareCropAnimated(input *imagick.MagickWand) (*imagick.MagickWand, error) {
// 	defer input.Destroy()
// 	coalesced := input.Clone().CoalesceImages()
// 	defer coalesced.Destroy()

// 	cropped := imagick.NewMagickWand()

// 	for i := 0; i < int(coalesced.GetNumberImages()); i++ {
// 		coalesced.SetIteratorIndex(i)
// 		frame := coalesced.GetImage()

// 		frame.SetImageDelay(coalesced.GetImageDelay())
// 		frame.SetImageDispose(coalesced.GetImageDispose())

// 		width := frame.GetImageWidth()
// 		height := frame.GetImageHeight()

// 		// Square crop size
// 		cropSize := width
// 		if height < width {
// 			cropSize = height
// 		}

// 		offsetX := int((width - cropSize) / 2)
// 		offsetY := int((height - cropSize) / 2)

// 		// Crop to center square
// 		if err := frame.CropImage(cropSize, cropSize, offsetX, offsetY); err != nil {
// 			cropped.Destroy()
// 			frame.Destroy()
// 			return nil, fmt.Errorf("crop image error: %w", err)
// 		}

// 		cropped.AddImage(frame)
// 		frame.Destroy()
// 	}

// 	// Optimize layers to reduce size (optional)
// 	if err := cropped.OptimizeImageLayers(); err != nil {
// 		return nil, fmt.Errorf("failed to optimize layers: %w", err)
// 	}

// 	return cropped, nil

// }

// func ResizeAnimated(input *imagick.MagickWand, maxSize uint) (*imagick.MagickWand, error) {
// 	defer input.Destroy()
// 	coalesced := input.Clone().CoalesceImages()
// 	defer coalesced.Destroy()

// 	resized := imagick.NewMagickWand()

// 	for i := 0; i < int(coalesced.GetNumberImages()); i++ {
// 		coalesced.SetIteratorIndex(i)
// 		frame := coalesced.GetImage()

// 		// Preserve animation metadata
// 		frame.SetImageDelay(coalesced.GetImageDelay())
// 		frame.SetImageDispose(coalesced.GetImageDispose())

// 		// Get original dimensions
// 		origW := float64(frame.GetImageWidth())
// 		origH := float64(frame.GetImageHeight())

// 		// Calculate new dimensions with aspect ratio preserved
// 		ratioW := float64(maxSize) / origW
// 		ratioH := float64(maxSize) / origH
// 		scale := min(ratioW, ratioH)

// 		newW := uint(origW * scale)
// 		newH := uint(origH * scale)

// 		// Resize the frame
// 		if err := frame.ResizeImage(newW, newH, imagick.FILTER_LANCZOS); err != nil {
// 			resized.Destroy()
// 			frame.Destroy()
// 			return nil, fmt.Errorf("resize image error: %w", err)
// 		}

// 		resized.AddImage(frame)
// 		frame.Destroy()
// 	}

// 	if err := resized.OptimizeImageLayers(); err != nil {
// 		return nil, fmt.Errorf("failed to optimize layers: %w", err)
// 	}

// 	return resized, nil
// }
