# Original filled-vector objects

The cylinder-38 segment is retained as `raw/stencil.bin`; these early segment
names are extraction labels, not exact visual scene names. It renders a large
cube and a 20-point faceted object before the later stencil-vector segment.

| Unit | Points | Edges | Faces | Motion bank | Source updates |
| --- | --- | --- | --- | --- | --- |
| Cube | `0xac38` | `0xac6a` | `0xac84` | `0xacaa` | 585 |
| Faceted solid | `0xacf0` | `0xad6a` | `0xadc0` | `0xae46` | 676 |

The geometry reader and signed 3D matrix are shared with the BOB scene. The
six-channel pose clock is shared with the credits. The original projector at
`0x9686` produces coordinates centered at (351,145); its display pointer skips
176 horizontal pixels and one source row, giving the visible (175,144) center.

The sine initializer at `0x936e` mirrors the supplied 512-word quarter into the
second quarter and negates the first half into the second half. The resulting
2,048 entries retain the source's exact integer values.

The entry code patches the VBL comparison operand at `0x9140` to **one**. Reading
only the unmodified render loop would incorrectly suggest two VBLs per update.
Both objects therefore advance at the authored 50 Hz.

Three 54-color ramps begin at `0xae7e`, `0xaf56` and `0xb02e`, each duplicated in
source memory. The cube colors ten-row groups; the second routine at `0x9286`
colors five-row groups. Separate counters cycle the ramps. Native rendering
stores every phase in one immutable atlas per object, then selects the phase
and material in a DCK triangle batch. Playback does not upload a new raster
image or allocate a texture per update.

`internal/demo/testdata/original-solid-projection.csv` contains 18 poses obtained
by executing the original sine initializer, matrix, transform and projector in
Ghidra's existing CPU emulator. Every vertex and corresponding cue-boundary
pose matches the Go port. The unused padding longword between transformed
points does not participate in projection. Native captures also cover the
interlude and each object's entrance, held motion and exit. Exact blitter pixel
parity and final global timing remain separate full-production checks.
