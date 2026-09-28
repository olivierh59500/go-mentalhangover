# Reference scene inventory

The supplied recording lasts 465.026 seconds at 50 frames/s. This initial map
used sparse observations; it now incorporates exact-frame samples and decoded
routine counts for the opening and Filled BOBs. Video extraction uses explicit
frame selection: an fps resampling filter can shift a contact-sheet timestamp
by half its sampling interval. Later boundaries remain provisional.

| Approximate recording time | Visible production unit | Candidate DCK reuse | Production-specific work |
| --- | --- | --- | --- |
| 0–12 s | Winged Scoopex emblem over stars | Fixed-step sprite field, image layers | Original bitplanes, per-bank palettes and music start |
| 12–30 s | Presents, title and coded-by cards | Bitmap pages, image layers, cue ranges | RGB12 flash/fades and source-centered lines |
| 30–70 s | Blue Slayer, Reward and Uncle Tom author objects | Bounded triangle batch, cue clock | Original word-sized projection and cue banks |
| 70–84 s | Follow the sign, sign raster sweep and Scoopex text card | Bitmap pages, masked raster overlay | Original 27 copper groups and 50-frame sweep |
| 84–123 s | Filled BOBs: scrolling, repeated cubes and pyramids | Scrolling facade, retained triangle batch | Binary controls, DMA-rounded lookup positions and raster colors |
| 123–153 s | Filled-vector introduction, cube and faceted solid | Retained triangle batch, palette atlas, cue clock | Original edge faces, mirrored sine bank and 54-phase rasters |
| 150–240 s | Stencil and patterned vector objects, textured panels and cube | Mesh rendering, masks, image layers | Exact stencil/texture rules and object sequence |
| 240–280 s | Greetings and member text pages over stars | Bitmap pages/scrolling, projected field | Font layout, messages and transitions |
| 280–345 s | Circular text over a raster-colored mountain silhouette | Text paths, raster materials and masking | Original circular layout and background bitplanes |
| 345–375 s | Contact text and moving blue objects/trails | Bitmap pages, projected sprites/histories | Authored positions, sprite bank and choreography |
| 375–420 s | Perspective blue glyph/object rows and final reminder | Projected geometry/text, timelines | Original font geometry, projection and movement |
| 420 s onward | Scoopex logo, perspective checkerboard and bouncing balls | Perspective checkerboard, projected balls | Original palettes, logo and exact floor/ball motion |

The star layer persists through several earlier units. Its simulation and depth
rules must be established from the original routine instead of selecting an
arbitrary generic distribution. Fifteen-second images do not establish complete
object lists or every transition. The final screen continues through the end of
the supplied movie; the original runtime determines its continuation behavior.
