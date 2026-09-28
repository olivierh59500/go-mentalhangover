# Native conversion verification

The renderer uses DCK **v1.0.0**, Go/Ebitengine **v2.9.11** and embedded data
decoded from the supplied disk. Neither the recording nor an emulator supplies
runtime frames. No analysis tool was installed for this work.

## Independent source evidence

Fixtures were produced by executing the original 68000 routines in the existing
Ghidra emulator. Go tests consume those outputs, not a second copy of the Go
formula. Graphics DMA calls are skipped only where the check targets geometry
or transport; raw point-plane checks execute the original bit operations.

| Unit | Independent comparison |
| --- | --- |
| Credits/sign | 44 original matrix/projector poses |
| Filled BOBs | Complete 1,971-update transport and 24 object poses |
| Large filled solids | 18 poses, including motion boundaries |
| Stencil units | 66 projector poses across all eight objects |
| Greeting star field | 15 exact original two-plane XOR raster hashes |
| Circle twist | 3,274 successful transport updates and all eleven profile rows |
| Contact spheres | 1,194 motion states and 1,426 visible blitter poses |
| Perspective | 3,080 lookup coordinates, 1,507 message updates and 12 exact point-plane hashes |
| Ending | 1,984 queue/palette states, 441 poses and 2,968 copper row-color pairs |

Source-specific boundary behavior is retained, including word arithmetic,
DMA-rounded BOB positions, byte text controls, the contact depth equality
comparison, the ending's size lookup one interval ahead, old-phase bounce
sampling and the palette endpoint's `0x0404` guard color.

GPU checks cover all 4,096 RGB12 colors through the six palette operations,
all eleven source-centered cards, outline parity/occlusion and repeated draws
without advancing clocks. The optimized final floor matches the independent
CPU raster pixel by pixel across reveal and palette cycles. It uploads a
2,176-byte row bank per tick and keeps its column mask on the GPU.

```sh
export GOWORK=off
go test ./...
go vet ./...
go test -tags mental_vector_rendercheck ./internal/demo
```

The last command requires an available graphical session. The ending also
passes 8,000 updates spanning complete palette and particle cycles.

## Recording alignment and music

Every authored effect is integrated in source order. Loading holds follow the
observed recording rather than local disk speed. Late entrances occur at
288.76 s (circle), 357.58 s (contact), 382.46 s (perspective), 418.96 s
(reminder) and 430.72 s (ending). Exact frame selection checks the prolonged
reminder and the first visible logo colors around 430.80 s.

The extracted 143,136-byte MOD matches the independent supplied soundtrack
byte for byte. DCK selects its module backend from the file; playback starts
at 11.30 s and loops with the indefinite ending. A complete 460-second PCM
render passes without decoder errors. Envelope comparisons at ten positions
from 20 to 430 s identify the same musical passages. The recorded replay has
slightly different timing: the best envelope offset grows from 11.33 to
11.98 s, a difference of approximately 0.16%. Decoder interpolation, filtering
and stereo mixing also affect waveform equality.

## Runtime and fidelity limits

A complete initial Pixel 10a run passed every production unit and continued
in the ending with its audio player active. Observed updates stayed close to
50 TPS; the 60 FPS display repeats PAL poses as needed. Final-build checks
exercise the optimized floor and corrected late adapters separately. The APK
targets ARM64, uses Java 17/SDK 36 and passes signature and 16 KiB packaging
alignment checks. Build outputs and verification captures remain local.

The conversion preserves original assets and checked integer geometry. GPU
triangle edges can differ from the original blitter's line/fill endpoint
conventions. The original beam-time random ball seeds are deterministic here,
so the exact final-ball arrangement differs from a particular recording.
These limits do not substitute prerecorded frames for native effects.
