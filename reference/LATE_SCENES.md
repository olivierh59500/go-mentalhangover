# Late-scene integration checkpoints

All three remaining executable segments now have native Go adapters:

- Cylinder 46 (`raw/circle-scroll.bin`) projects 130 source XY points using
  the original reciprocal table, depth queue and sixteen sphere sizes. The
  contact text is retained from eight 16-pixel hardware-sprite columns. The
  sphere and text palettes have independent fades. The complete phase program
  currently consumes 1,194 source updates.
- Cylinder 48 (`raw/final-reminder.bin`) reuses the same packed outline-font
  reader as the circle-twist. Its eleven-row, 280-column perspective lookup
  preserves signed integer division. Eight slots are painted through DCK's
  scrolling facade, with a source-colored blue row palette and separate point
  layer. The finite source message and fades currently consume 1,677 updates.
- Cylinder 49 (`raw/checkerboard.bin`) retains the 320 by 87 four-plane logo,
  prepared fifth-plane column mask, ninety-color floor ramp, 24-entry depth
  queue, byte bounce table and six sphere materials with pre-rendered sizes.
  DCK triangle batches draw the sphere copies. The ending continues running.

The lower finale display loads pointers at row `0x87` but enables bitplane DMA
at row `0xab`; bitmap advancement begins at the following line. Native floor
sampling and ball placement retain that 37-row distinction. Color wrapping
is bounded in Go, avoiding an out-of-bank read at the source palette boundary.
The source mixes beam timing into new ball seeds; the native replay uses a
fixed seed so captures remain reproducible.

Native capture checkpoints cover entrances, holds, exits, perspective text,
floor reveal and multiple ending cycles. Pure checks cover complete text/phase
programs, all final asset dimensions and 8,000 ending updates. Final source CPU
comparisons for the late adapters, whole-production timing/color calibration
and complete audio/runtime verification are still pending. These checks must
be completed before claiming finished conversion or full visual fidelity.
