# Original stencil-vector sequence

The cylinder-41 segment at `0x8000` makes eight separate calls into its renderer.
Their original order is retained: gold arrow, second arrow pass, yellow cube,
brown faceted solid, paired boxes, patterned pyramid, hollow frame and color
block. The complete sequence takes 4,365 PAL ticks, or 87.30 seconds, excluding
the preceding loading/interlude hold.

The brown object and hollow frame patch the VBL operand at `0x830e` to two.
They update at 25 Hz while the other units update at 50 Hz. Stars and module
playback continue on the 50 Hz production clock. Drawing never advances a pose.

Unlike the earlier edge-indexed faces, these faces contain repeated closing
point indices. A byte with its high bit set separates parity contours inside
one face. Native rendering submits each face through a DCK even-odd batch,
preserving counters and openings, including the central hole of the frame.

The 360-word degree sine table starts at `0xa458`. Angles retain the original
word offsets and one-step wrapping at 720 bytes. Matrix products use signed
high-word truncation with seven fractional shift bits. Projection keeps the
unhalved camera depth, 128-scale translations and (256,100) source origin.
The display pointer skips 80 horizontal pixels and three source rows; copper
row advancement starts 30 rows below the opening's visible origin.

Each original pattern bank is a 224 by 32 pixel strip, containing seven 32 by
32 materials with three interleaved planes. The renderer repeats these motifs
and centers their screen-space sampling on each projected face. Native texture
repeat handles wrapping without a per-frame copy or texture allocation. Depth
selects one of the 128 original nibble-truncated palettes.

`internal/demo/testdata/original-pattern-projection.csv` records 66 poses from
all eight objects, obtained by executing the source matrix, transform and
projection routines at `0x84ac`, `0x8682` and `0x876e`. Every checked vertex
matches the Go port. Native tests also exercise the actual half-rate effects,
held projections and repeated drawing. Captures cover all eight original
materials/objects. Exact line-blitter coverage, one-pixel texture alignment
and final global cue alignment remain full-production comparison work.
