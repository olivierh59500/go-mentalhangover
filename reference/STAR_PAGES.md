# Steered stars and greetings

The cylinder-39 segment, initially retained as `raw/textured-cube.bin`, contains
a 260-point steered field and two complete text pages. Source descriptors at
`0x973c` and `0x98ae` contain the greetings and member list; no message is invented
or replaced with placeholder text. The original small font starts at `0xa668`,
with its 59 advances at `0x96c6`. It uses two interleaved planes, fourteen visible
rows and sixteen-pixel cells. DCK's page renderer prepares each page once.

The source display starts ten rows earlier than the opening, so page placement
retains that offset in the common native viewport. Its copper clears sprite DMA;
the earlier hardware-star adapter continues its clock but is not drawn here.

Steering angles use the degree sine table at `0xa398`. Three original matrix
products provide signed word velocities; their high-word results have eleven,
eleven and nine shift bits. XY coordinates wrap at 512, depth at 2,048. The
projection table contains `262144 / (depth + 10)` with integer truncation.
The source draws into two interleaved planes using XOR and depth thresholds.

Native stepping clears only the previously touched pixels and folds the 260
points into the same two-bit mask. DCK submits the remaining pixels in a single
retained triangle batch. It preserves overlapping-particle cancellations
without a full-frame upload, and stepping allocates no memory.

The original point fade, two page fades and holds consume 1,747 updates, or
34.94 seconds. The point and page palettes use separate 32-step nibble arithmetic.
Page holds use level 31; the point field reaches level 32. Changes in steering
continue throughout page transitions and empty intervals.

`internal/demo/testdata/original-star-pages.csv` records fifteen independent
executions of the original reciprocal initializer, steering and pixel routine
at `0x95e6`, `0x94d0` and `0x941e`. All updated offsets and full two-plane pixel
hashes match the Go port, including signed wrapping and particle cancellations.
Native tests also check cached page rendering and drawing without advancing
the field. This establishes the checked point bitmaps; final video alignment
and the remaining production units are still being completed.
