package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	url           = "https://api.replicate.com/v1/predictions"
	version       = "7762fd07cf82c948538e41f63f77d685e02b063e37e496e96eefd46c929f9bdc"
	token         = "3e352b505e7f33370501180343d5443169fe1cce"
	pubIp         = "2607:fa49:5602:9600::2bc8"
	hookAuthority = "https://fa17-2607-fa49-5602-9600-00-2bc8.ngrok-free.app"
	hookUrl       = hookAuthority + "/generate-media-hook"

	// url     = "https://api.replicate.com/v1/models/stability-ai/stable-diffusion-3/predictions"
)

type hookMsg struct {
	id   string
	data []byte
}

type hookMan struct {
	mch chan *hookMsg
	mm  map[string]chan []byte
	// add     chan string
	timeout chan string
}

func (hm *hookMan) Run() {
	log.Println("Running Hook Manager!")
	for {
		select {
		// case a := <-hm.add:
		// 	log.Println("adding chan for id:", a)
		// 	hm.mm[a] = make(chan []byte)
		case s := <-hm.timeout:
			log.Println("timeout request for id:", s)
			if ch := hm.mm[s]; ch != nil {
				log.Println("timeing out id:", s)
				close(ch)
				delete(hm.mm, s)
			}
		case m := <-hm.mch:
			log.Println("message request for id:", m.id)
			if ch := hm.mm[m.id]; ch != nil {
				log.Println("sending the message for id:", m.id)
				ch <- m.data
				close(ch)
				delete(hm.mm, m.id)
			}
		}
	}
}

var HookManager = &hookMan{
	mch: make(chan *hookMsg),
	mm:  make(map[string]chan []byte),
	// add:     make(chan string),
	timeout: make(chan string),
}

func HandleGenerateMediaHook(w http.ResponseWriter, r *http.Request) {
	var rjsn map[string]any
	if err := json.NewDecoder(r.Body).Decode(&rjsn); err != nil {
		log.Println("HandleGenerateMediaHook error decoding req:", err)
		return
	}

	// mi, _ := json.MarshalIndent(rjsn, "", "    ")
	// log.Println("HandleGenerateMediaCallback response:\n", string(mi))

	status, ok := rjsn["status"].(string)
	id, ok2 := rjsn["id"].(string)
	if !ok || !ok2 {
		log.Println("HandleGenerateMediaHook status is not a key")
		return
	}

	if status == "succeeded" {
		log.Println("HandleGenerateMediaHook success")
		// log.Printf("type of rjson['output'] = %T\n", rjsn["output"])
		output, ok := rjsn["output"].([]interface{})
		if !ok {
			log.Println("HandleGenerateMediaHook output is not an array")
			return
		}

		if len(output) != 1 {
			log.Println("HandleGenerateMediaHook output is not of length 1")
			return
		}

		imurl, ok := output[0].(string)
		if !ok {
			log.Println("HandleGenerateMediaHook imurl is not a string")
			return
		}

		req, err := http.NewRequest("GET", imurl, nil)
		// req.Header.Add("Authorization", "Bearer "+token)
		if err != nil {
			log.Println("HandleGenerateMediaHook error building req:", err)
			return
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("HandleGenerateMediaHook error doing req:", err)
			return
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("HandleGenerateMediaHook error reading res body:", err)
			return
		}

		HookManager.mch <- &hookMsg{id: id, data: data}
	} else {
		log.Println("HandleGenerateMediaHook was not a success")
		HookManager.mch <- &hookMsg{id: id, data: nil}
	}
}

func HandleGenerateMedia(w http.ResponseWriter, r *http.Request) {
	var genReq struct {
		MediaId       string   `msgpack:"id"`
		MediaMetadata []byte   `msgpack:"media"`
		Prompts       []string `msgpack:"prompts"`
	}

	if err := msgpack.NewDecoder(r.Body).Decode(&genReq); err != nil {
		w.WriteHeader(500)
		log.Println("error decoding body json:", err)
		return
	}

	// log.Println("input:", genReq.Prompts, "id:",
	// 	genReq.MediaId, "media:", genReq.MediaMetadata)

	jsn := map[string]any{
		"version":               version,
		"webhook":               hookUrl,
		"webhook_events_filter": []string{"completed"},
		"input": map[string]any{
			"width":               768,
			"height":              768,
			"refine":              "expert_ensemble_refiner",
			"scheduler":           "K_EULER",
			"lora_scale":          0.6,
			"num_outputs":         1,
			"guidance_scale":      7.5,
			"apply_watermark":     false,
			"high_noise_frac":     0.8,
			"negative_prompt":     "",
			"prompt_strength":     0.8,
			"num_inference_steps": 12,

			// "aspect_ratio":    "1:1",
			// "cfg":             3.5,
			// "output_format":   "jpg",
			// "output_quality":  50,
			// "steps":           14,
			// "prompt_strength": 0.85,
		},
	}
	for _, prompt := range genReq.Prompts {
		log.Println("trying with prompt=", prompt)
		jsn["input"].(map[string]any)["prompt"] = prompt
		b, err := json.Marshal(jsn)
		if err != nil {
			w.WriteHeader(500)
			log.Println("error encoding json:", err)
			return
		}
		// log.Println("prettyInput:\n", string(b))

		req, err := http.NewRequest("POST", url, bytes.NewReader(b))
		if err != nil {
			w.WriteHeader(500)
			log.Println("error building request:", err)
			return
		}

		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(500)
			log.Println("error doing request:", err)
			return
		}

		if res.StatusCode == 201 {
			var rjsn map[string]any
			if err = json.NewDecoder(res.Body).Decode(&rjsn); err != nil {
				log.Println("error decoding response:", rjsn)
				w.WriteHeader(500)
				return
			}

			// prettyRes, _ := json.MarshalIndent(rjsn, "", "    ")
			// log.Println("response:\n", string(prettyRes))

			id, ok := rjsn["id"].(string)
			if !ok {
				log.Println("response id is not a string")
				w.WriteHeader(500)
				return
			}

			log.Println("adding chan for id:", id)
			HookManager.mm[id] = make(chan []byte)
			// HookManager.add <- id

			timer := time.NewTimer(time.Second * 30)
			select {
			case <-timer.C:
				HookManager.timeout <- id
			case data := <-HookManager.mm[id]:
				if data == nil {
					log.Println("GenerateMedia failed, trying again")
					continue
				} else {
					log.Println("GenerateMedia success, got the media")
					l := 2 + len(genReq.MediaMetadata) + len(data)
					buf := make([]byte, 0, l)
					wb := bytes.NewBuffer(buf)
					metadataLen := uint16(len(genReq.MediaMetadata))
					binary.Write(wb, binary.BigEndian, metadataLen)
					wb.Write(genReq.MediaMetadata)
					wb.Write(data)
					// log.Printf("wb.Bytes() len = %d, buf len = %d",
					// 	len(wb.Bytes()), len(buf))
					// err = WriteMedia2(genReq.MediaId, true, buf)
					// log.Println("GenerateMedia, writing media to disk")
					err = WriteMedia(genReq.MediaId, false, wb)
					if err != nil {
						log.Println("GenerateMedia failed write")
						w.WriteHeader(500)
						return
					}

					binary.Write(w, binary.BigEndian, uint16(len(prompt)))
					w.Write([]byte(prompt))
					w.Write(data)
				}
			}

		} else {
			log.Println("GenerateMedia problem making generation request, closing")
			w.WriteHeader(500)
			return
		}
	}
}
