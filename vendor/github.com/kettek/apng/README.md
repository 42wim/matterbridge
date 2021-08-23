# APNG golang library
This `apng` package provides methods for decoding and encoding APNG files. It is based upon the work in the official "image/png" package.

See [apngr](https://github.com/kettek/apngr) for an APNG extraction and combination tool using this library.

**NOTE**: The decoder should work for most anything you throw at it. Malformed PNGs should result in an error message. The encoder currently doesn't handle differences of Image formats and similar and has not been tested as thoroughly.

If a regular PNG file is read, the first Frame of the APNG returned by `DecodeAll(*File)` will be the PNG data.

## Types
### APNG
The APNG type contains the frames of a decoded `.apng` file, along with any important properties. It may also be created and used for Encoding.

| Signature                 | Description                   |
|---------------------------|-------------------------------|
| Frames []Frame            | The stored frames of the APNG.|
| LoopCount uint            | The number of times an animation should be restarted during display. A value of 0 means to loop forever.   |

### Frame
The Frame type contains an individual frame of an APNG. The following table provides the important properties and methods.

| Signature                 | Description      |
|---------------------------|------------------|
| Image image.Image           | Frame image data. |
| IsDefault bool            | Indicates if this frame is a default image that should not be included as part of the animation frames. May only be true for the first Frame. |
| XOffset int               | Returns the x offset of the frame. |
| YOffset int               | Returns the y offset of the frame. |
| DelayNumerator int        | Returns the delay numerator.       |
| DelayDenominator int      | Returns the delay denominator.     |
| DisposeOp byte           | Returns the frame disposal operation. May be `apng.DISPOSE_OP_NONE`, `apng.DISPOSE_OP_BACKGROUND`, or `apng.DISPOSE_OP_PREVIOUS`. See the [APNG Specification](https://wiki.mozilla.org/APNG_Specification#.60fcTL.60:_The_Frame_Control_Chunk) for more information. |
| BlendOp byte              | Returns the frame blending operation. May be `apng.BLEND_OP_SOURCE` or `apng.BLEND_OP_OVER`. See the [APNG Specification](https://wiki.mozilla.org/APNG_Specification#.60fcTL.60:_The_Frame_Control_Chunk) for more information. |

## Methods
### DecodeAll(io.Reader) (APNG, error)
This method returns an APNG type containing the frames and associated data within the passed file.

### Example
```go
package main

import (
  "github.com/kettek/apng"
  "os"
  "log"
)

func main() {
  // Open our animated PNG file
  f, err := os.Open("animation.png")
  if err != nil {
    panic(err)
  }
  defer f.Close()
  // Decode all frames into an APNG
  a, err := apng.DecodeAll(f)
  if err != nil {
    panic(err)
  }
  // Print some information on the APNG
  log.Printf("Found %d frames\n", len(a.Frames))
  for i, frame := range a.Frames {
    b := frame.Image.Bounds()
    log.Printf("Frame %d: %dx%d\n", i, b.Max.X, b.Max.Y)
  }
}

```

### Decode(io.Reader) (image.Image, error)
This method returns the Image of the default frame of an APNG file.

### Encode(io.Writer, APNG) error
This method writes the passed APNG object to the given io.Writer as an APNG binary file.

### Example
```go
package main

import (
  "github.com/kettek/apng"
  "image/png"
  "os"
)

func main() {
  // Define our variables
  output := "animation.png"
  images := [4]string{"0.png", "1.png", "2.png", "3.png"}
  a := apng.APNG{
    Frames: make([]apng.Frame, len(images)),
  }
  // Open our file for writing
  out, err := os.Create(output)
  if err != nil {
    panic(err)
  }
  defer out.Close()
  // Assign each decoded PNG's Image to the appropriate Frame Image
  for i, s := range images {
    in, err := os.Open(s)
    if err != nil {
      panic(err)
    }
    defer in.Close()
    m, err := png.Decode(in)
    if err != nil {
      panic(err)
    }
    a.Frames[i].Image = m
  }
  // Write APNG to our output file
  apng.Encode(out, a)
}
```
