package libtgsconverter

import "bytes"

import "image"
import "image/color"
import "image/gif"

type togif struct {
	gif gif.GIF
	images []image.Image
	prev_frame *image.RGBA
}

func(to_gif *togif) init(w uint, h uint, options ConverterOptions) {
	to_gif.gif.Config.Width = int(w)
	to_gif.gif.Config.Height = int(h)
}

func(to_gif *togif) SupportsAnimation() bool {
	return true
}

func (to_gif *togif) AddFrame(image *image.RGBA, fps uint) error {
	var fps_int = int(1.0 / float32(fps) * 100.)
	if to_gif.prev_frame != nil && sameImage(to_gif.prev_frame, image) {
		to_gif.gif.Delay[len(to_gif.gif.Delay) - 1] += fps_int
		return nil
	}
	to_gif.gif.Image = append(to_gif.gif.Image, nil)
	to_gif.gif.Delay = append(to_gif.gif.Delay, fps_int)
	to_gif.gif.Disposal = append(to_gif.gif.Disposal, gif.DisposalBackground)
	to_gif.images = append(to_gif.images, image)
	to_gif.prev_frame = image
	return nil
}

func (to_gif *togif) Result() []byte {
	q := medianCutQuantizer{mode, nil, false}
	p := q.quantizeMultiple(make([]color.Color, 0, 256), to_gif.images)
	// Add transparent entry finally
	var trans_idx uint8 = 0
	if q.reserveTransparent {
		trans_idx = uint8(len(p))
	}
	var id_map = make(map[uint32]uint8)
	for i, img := range to_gif.images {
		pi := image.NewPaletted(img.Bounds(), p)
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := img.At(x, y)
				cr, cg, cb, ca := c.RGBA()
				cid := (cr >> 8) << 16 | cg | (cb >> 8)
				if q.reserveTransparent && ca == 0 {
					pi.Pix[pi.PixOffset(x, y)] = trans_idx
				} else if val, ok := id_map[cid]; ok {
					pi.Pix[pi.PixOffset(x, y)] = val
				} else {
					val := uint8(p.Index(c))
					pi.Pix[pi.PixOffset(x, y)] = val
					id_map[cid] = val
				}
			}
		}
		to_gif.gif.Image[i] = pi
	}
	if q.reserveTransparent {
		p = append(p, color.RGBA{0, 0, 0, 0})
	}
	for _, img := range to_gif.gif.Image {
		img.Palette = p
	}
	to_gif.gif.Config.ColorModel = p
	var data []byte
	w := bytes.NewBuffer(data)
	err := gif.EncodeAll(w, &to_gif.gif)
	if err != nil {
		return nil
	}
	return w.Bytes()
}
