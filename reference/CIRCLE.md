# Original filled circle-twist

The cylinder-44 segment, retained as `raw/greetings.bin`, renders a circular
outline-font scrolling over the mountain mask at `0x9fe8`. The mask is one
352 by 200 pixel plane. The 49 copper colors at `0x9868` each cover two rows.
DCK image surfaces retain both assets, and the original 33-step fades use
nibble-truncated RGB12 arithmetic.

The 59 outline-font pointers start at `0x9958`. Each glyph stores an edge
count followed by coordinate pairs; a high-bit marker separates closed
contours. The Go reader retains concave outlines and counters. Eight visible
slots are painted through `scrolling.New` and one DCK parity batch. The same
source outline data can be mapped by another production-specific point function.

The generated polar table contains 64 radius rows and 256 angle positions.
Its radii are **0,2,...126**, because the original initializer increments twice
per row. The coordinate byte wave at `0x9d12` has amplitude 99. A separate
127-amplitude wave at `0x9e12` selects the base radius. Eleven radial profiles
deform the glyph rows independently.

The profile's word offset retains its `0xff` high byte as its low phase byte
wraps, selecting the preceding coordinate wave. Intermediate negative products
also retain high-word residues before a long shift. Their addresses alias the
same polar table on the Amiga 500's 24-bit bus. The native adapter preserves the
resulting low-address coordinates rather than replacing them with a smooth
floating-point sine approximation.

The binary message starts at `0x9f12`. `|` supplies a pause; `>` supplies five
bytes controlling radius phase step, radius factor, row-profile step and both
phase bytes. Commands apply after the current profile calculation. The exact
look-ahead transport is retained, including its startup and byte wrap. The
mountain plane occludes the lower text and its two-row dark shadow.

`internal/demo/testdata/original-circle-clock.csv` records all 3,274 successful
executions of the original clock/profile routine at `0x91e6`, followed by its
end redirect. Ghidra's 32-bit CPU model uses an explicit alias of the generated
table to reproduce the 24-bit address bus. Only the line-drawing subroutine is
replaced with its register-preserving return in this analysis fixture.
Message pointers, angle byte, pauses, five control values and every profile
row match Go. Including the end redirect and two 33-step fades, the native
unit consumes 3,341 updates (66.82 seconds).

Native checks cover repeated draws without changing profiles and exact
occlusion by the decoded mountain mask. Line/fill blitter edge parity,
video color calibration and final global alignment remain full-production
comparison work; no claim of complete pixel identity is made here.
