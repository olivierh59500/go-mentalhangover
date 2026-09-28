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
sampling and ball placement retain that 37-row distinction. The palette's
permitted endpoint aliases the original size table's first word (`0x0404`);
Go retains that color explicitly without reading beyond its palette slice.
The source mixes beam timing into new ball seeds; the native replay uses a
fixed seed so captures remain reproducible.

Independent executions of the original CPU routines now cover all 1,194
contact motion states and 1,426 visible sphere poses, all 3,080 perspective
coordinates and 1,507 successful message updates, and twelve exact hashes of
the perspective point bitplanes. Coincident points retain the source's OR
operation. Ending fixtures cover 1,984 queue/palette states, 441 sphere poses
and 2,968 copper row-color pairs; only beam-time randomness is substituted.
These checks found and fixed the contact queue's equality threshold, the
ending's size lookup one interval ahead and its old-phase bounce sampling.

The finale retains its column mask on the GPU and uploads only a two-color row
bank: 2,176 bytes per update instead of 382,976. GPU checks compare the result
pixel by pixel against the independent CPU raster across reveal and color
cycles. Repeated draws of all late units preserve their clocks and pixels.
Pure checks also cover 8,000 ending updates. A complete initial Pixel 10a run
reached the indefinite ending with audio active and approximately 50 TPS.

The loading/reminder holds are now aligned to exact-frame recording samples.
The updated Android build also renders the optimized ending at approximately
50 TPS. [Whole-production checks and fidelity limits](VERIFICATION.md) document
the native GPU rendering and the measured music timing difference.
