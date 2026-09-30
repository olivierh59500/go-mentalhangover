# Mental Hangover Go

Native Go/Ebitengine conversion of the supplied Amiga Mental Hangover
production, using **Demo Construction Kit v1.0.10** and production-specific code.

## Overview

The port follows the original 50 Hz PAL sequence. DCK composes the effects
and plays the four-channel ProTracker soundtrack; original-specific formats,
materials and choreography are implemented in Go. The supplied recording was
used to check scene timing and motion. Runtime rendering is native Ebitengine.

## Implementation and checks

The custom loader's eleven cylinder transfers are reproduced in Go, including
its 32-bit XOR.
The extracted soundtrack matches the independently supplied module exactly.
The native director assembles the entire production sequence, from the eagle and
title through the author credits, "Follow the sign", the sign's scrolling
copper raster and the six-line Scoopex interlude, followed by the complete
Filled BOBs scrolling and its 35 projected cubes/pyramids, then the two large
filled-vector objects and their moving raster colors, then all eight stencil
units with the original tiled motifs, followed by the steered star field and
the complete greetings/member pages and the entire circular outline-font twist
over the original mountain mask and copper colors, the contact spheres,
perspective outline text, final reminder and the animated checkerboard/balls.
The ending continues running after its initial sequence. It uses the original
text descriptors, integer-centered font, copper palette operations and vector
motion programs. The module starts at 11.30 seconds, calibrated against the
supplied recording; this includes the recording's initial disk-loading wait.
The BOBs retain the binary commands embedded in the source message, gradual
speed changes, text pauses and independent object clocks. All 1,971 text
updates and 24 solid-object projections match executions of the original CPU
routines. Eighteen additional poses of the large solids, including cue
boundaries, match their original projector. The 66 checked stencil poses match
the original CPU
projector; the two half-rate units hold their poses between PAL updates.
Fifteen steered-star views match the original two-plane XOR output exactly.
Those points use a bounded sparse raster and one DCK batch; pages are drawn
once into retained surfaces rather than redrawing hundreds of glyphs per tick.
The circle's complete transport and eleven deformation rows match 3,274
successful original CPU updates, including its pause and form-change commands.
The late units also retain their independently verified sphere projections,
perspective lookup, point-plane OR operations, depth queue and copper colors.
The reminder stays visible through the final disk-transfer hold, and the ending
begins at 430.72 seconds. Native captures cover the entire sequence. An earlier
complete Pixel run reached the indefinite ending with audio active and
approximately 50 updates/s. This is a native reconstruction: GPU polygon edges and the
deterministic final-ball seeds differ from cycle-accurate Amiga emulation.
Detailed numeric evidence and measured audio differences are documented in
[verification](reference/VERIFICATION.md).

The circular and perspective text now use DCK's `font.ContourBank`, owned
`scrolling.Mode.Contours` and authored `GlyphWindow`. The shared `ByteWindow`
handles signed-byte positions, command lookahead, pauses and completion; shared
`TablePolar`/`RationalGrid` maps preserve original integer projection and wrapping.
The demo supplies artwork, wave/profile tables, coefficients, scene fades and
layer order. Its local polygon painters and dummy glyph setup are removed.
All 699 sampled complete RGBA frames match the preceding renderer over 24,001
drawn updates, including 69 samples in those two scenes. Original CPU fixtures
for their transport, profiles and projections continue to pass. No intermediate
image or shader is added by this migration.

The independent asset preview renders the original four-plane eagle, interleaved
three-plane serif font and 135 original star headers through DCK. The stars
retain their three fixed horizontal velocities and byte wraps at 50 Hz.

```sh
export GOWORK=off
go run ./cmd/extract -disk /path/to/Scoopex-MentalHangover.adf -module /path/to/madness.mod
go run ./cmd/mentalhangover
go run ./cmd/mentalhangover -mute -capture captures/sequence
go run ./cmd/preview
go run ./cmd/preview -font
go run ./cmd/preview -title
go run ./cmd/preview -vector slayer
go run ./cmd/preview -mute -capture captures/intro-assets
```

Desktop controls: **F** toggles fullscreen; **Escape** exits. The demo uses the
embedded production assets and needs no reference disk or video at runtime.

`cmd/preview` is an asset-validation host. The full director includes all
decoded effect blocks.

The Slayer, Reward, Uncle Tom and sign objects now retain their original
word-sized motion programs and closed polygon banks. Their 44 checked poses
match an independent execution of the original 68000 matrix/projection code.
The title's five-plane bitmap and palette are also decoded directly. These
models are integrated into the director. [Numeric and rendering evidence](reference/VECTORS.md)

The native GPU checks cover all 4,096 RGB12 colors at all 293 mode/level samples
through the six palette operations, all eleven text cards against independent
integer pen placement,
contour parity, and repeated draws without advancing the source
clock or changing a vector pose.

The six copper operations now use DCK's `composite.QuantizedColor`; the production
supplies its grid, threshold, integer ratios and operation order. All 699 sampled
full frames match the previous renderer across 24,001 drawn ticks and 41 units.
The shared pass owns one shader without an additional image surface. Two muted
native
CPU submission runs per implementation measured mean draws of 37.82–38.63
microseconds before and 38.27–39.70 after; those timings exclude GPU completion
and readback. Reproduce frame fingerprints and optional submission timings with
`go run ./cmd/checkframes -output captures/color-reuse.json -timing`.

Author vectors, BOBs, filled solids and patterned, circular and perspective
contours use DCK's `render.Batch.Fan`. The production retains its word-sized
projection, clipping, UVs and original coordinate programs; the shared builder
maps each vertex once and submits the same ordered triangles. All 699 sampled
frames still match across the complete 24,001-tick traversal. Immediate contour
drawing adds no working image or GPU pass.

The finale uses DCK's `composite.PaletteGrid` in one-column row mode. Its original
column mask and two RGB12 colors per row drive the same checkerboard material;
the shared pass owns the 2,176-byte row bank and supports continuous or binary
control with configurable channels and offsets. All 699 sampled full frames
still match through 24,001 drawn ticks. The independent CPU floor raster also
matches at the source's palette/ball transition ticks.

[Scene inventory](SCENES.md) · [Source manifest](reference/sources.json)

[Android build and Pixel verification](ANDROID.md):
`./scripts/run-android.sh` builds, installs and launches the same production.

## Video export

The recording command uses DCK's synchronized canvas/audio exporter. Its
default eight-minute recording includes the complete sequence and the looping
ending, at 704 by 544 pixels and the original 50 Hz. It captures the demo's
canvas and module audio only. The PNG poster defaults to the perspective scroll.

```sh
GOWORK=off go run ./cmd/video -output recordings/mental-hangover.mp4
```

`-duration`, `-poster-at`, `-crf` and `-preset` configure the export. A positive
duration is required because the original ending continues indefinitely.
Existing recordings are never overwritten. Generated MP4s, posters and
recording reports are produced alongside the video.
