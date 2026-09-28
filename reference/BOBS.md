# Original Filled BOBs reconstruction

The cylinder-36 segment, retained as `raw/filled-vector.bin`, starts its BOB
scene at `0x9038`. It does not render a generic ribbon: it projects a source
cube or pyramid and copies the result to positions from two lookup tables.
DCK `render.Batch` submits those copies in one retained batch with the three
original copper color ramps selected by destination row. No surface or sprite
texture is allocated for each copy or frame.

| Object | Point bank | Edge bank | Face bank | Points / edges / faces |
| --- | --- | --- | --- | --- |
| Cube | `0xab30` | `0xab62` | `0xab7c` | 8 / 12 / 6 |
| Pyramid | `0xaba2` | `0xabc2` | `0xabd4` | 5 / 8 / 5 |

Faces reference edges, rather than a repeated list of point indices. The Go
reader recovers the authored boundary and first-edge orientation before the
visibility test. Projection uses the original 2,048-word sine bank, signed
high-word products, word wrapping, camera approach and signed word division.
The solid matrix changes the order of one intermediate truncated product from
the credit matrix; that distinction is retained. Projected faces are clipped
to the original 48 by 32 pixel copy area, including frames where a vertex is
above the source tile.

The X lookup at `0xbed8` stores a DMA byte offset and fine shift per entry.
Chip DMA discards bit zero of word addresses. Applying that rounding before
conversion to pixel positions prevents eight-pixel jumps. The Y lookup starts
at `0xc2d8` and uses offsets in the 88-byte interleaved destination rows.

## Scrolling and binary controls

The message starts at `0xac86`. Its font starts at `0x9b58`: sixty stored
16-pixel cells, two interleaved planes and fourteen visible rows. Supported
character advances come from the 59-word table at `0x9ae2`, plus one.
DCK `scrolling.New` draws these glyphs with the original insertion positions.

- `>` changes the target speed. The actual speed approaches it one word unit
  per active update; transport uses half that actual speed in pixels.
- `|` pauses text movement for multiples of 50 frames. Object motion continues.
- `b` consumes fourteen binary bytes specifying count, two phase velocities,
  two copy spacings, phases, three angle deltas, three initial angles and model.
  Embedded zero bytes are data, not string terminators.
- `a` approaches camera depth 1,100; `d` approaches depth 4,100. The object pose
  is retained throughout those changes.

## Independent evidence and limits

`internal/demo/testdata/original-bob-clock.csv` records all 1,971 updates from
executing the original routine at `0x98a0` in Ghidra's existing CPU emulator.
Tests compare message pointers, proportional remainders, actual/target speed,
pauses, camera targets, copy counts, spacing and angle deltas. The original
return redirect ends the scene on update 1,971, for 39.42 seconds at 50 Hz.

`original-bob-projection.csv` records 24 poses obtained by executing the original
matrix, transform and projector at `0x91dc`, `0x934e` and `0x947a`. Every checked
vertex matches the Go calculation. Pure tests also cover the complete source
model topology and allocation-free stepping. Native checks compare text pixels
to direct source-glyph placement and check that repeated drawing keeps clocks
and pixels unchanged.

These checks establish the numeric routines and the native adapters. They do
not establish exact pixel parity with the original Amiga line/fill blitter or
all scene-to-music boundaries. Boot and disk-loading waits remain calibrated
to the supplied recording. Later production units remain in progress.
