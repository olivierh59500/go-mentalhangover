# Original credit-vector reconstruction

The author/ribbon unit starts at Amiga address `0x9000`. Its call sequence
places the following banks after the corresponding source text cards:

| Object | Points | Faces | Initial state and cue bank | Points / contours | Motion updates |
| --- | --- | --- | --- | --- | --- |
| Slayer | `0xac88` | `0xad6e` | `0xadc2` | 57 / 8 | 532 |
| Reward | `0xaab2` | `0xabe4` | `0xac50` | 76 / 10 | 570 |
| Uncle Tom | `0xadfa` | `0xaebc` | `0xaf0a` | 48 / 9 | 544 |
| Sign | `0xa972` | `0xaa48` | `0xaa88` | 53 / 3 | 256 |

Points have signed 16-bit X/Y coordinates. Each contour has two header bytes,
then its byte-sized indices including a repeated closing index. The next
contour follows immediately; only the end of the complete bank has alignment
padding. This distinction preserves the packed Slayer and Uncle Tom data.

The renderer XORs contour edges into one bitplane before applying an even-odd
fill. The native adapter therefore submits all triangle fans through one DCK
`render.Batch` with Ebitengine's even-odd fill rule. Concave letters, counters
inside letters and the sign's crossing edges retain their parity.

Six signed word-sized channels contain three angle-table offsets, X/Y offsets
and camera depth. Cues change six word deltas on their boundary update and keep
the preceding pose. The 2,048-word sine bank is retained exactly. Intermediate
matrix operations keep the original signed high-word truncation and projection
keeps signed word division, overflow behavior and the (175,136) origin.

## Independent numeric evidence

`internal/demo/testdata/original-projection.csv` contains coordinates obtained
by executing the original routines at `0x9326`, `0x947e` and `0x94c4` with
Ghidra's existing Motorola CPU emulator. The fixture covers 44 poses across
the four source objects, including cue boundaries and late motion. Every
vertex matches the Go projection exactly. This proves the checked numeric
poses, not pixel identity of the original blitter fill or the complete scene.

Pure tests check all cue lengths, packed contour closure, source indices,
signed word wrapping, retained boundary poses and allocation-free stepping.
The native `mental_vector_rendercheck` separately checks a concave contour and
an inner hole against a direct pixel-mask model.

The five-plane Mental Hangover title is a separate source bitmap beginning at
`0x12058`. Its rows interleave five 44-byte planes and the copper enables 163
rows. The 22 source palette words begin at `0x99b0`.
