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
| 153–247 s | Stencil interlude, two arrow passes, yellow cube, brown facets, paired boxes, pyramid, hollow frame and color block | Triangle batches with parity masks, repeating materials, cue clocks | Original packed contours, degree table, 128-level depth palette and 25/50 Hz cadences |
| About 250–285 s | Steered point field with greetings and member text pages | Cached bitmap pages, retained pixel batch | Original reciprocal projection, XOR collisions, steering and separate 32-step palettes |
| About 287–354 s | Circular text over a raster-colored mountain silhouette | Scrolling facade with outline painter, parity batch, retained image layers | Original byte controls, signed wave addressing, polar lookup and mountain occlusion |
| About 356–380 s | Contact text and projected blue spheres | Projected sprite batches, held image layers | Authored positions, size bank, steering and hardware text columns |
| About 382–420 s | Perspective blue text and points | Scrolling outline painter, projected geometry, pixel batch | Original font, perspective lookup and message transport |
| About 420–430 s | Final reminder card and disk-loading interval | Bitmap pages, cue ranges | Original flash/fades and last loader hold |
| About 430 s onward | Scoopex logo, perspective checkerboard and bouncing balls | Perspective checkerboard, projected balls | Original palettes, logo and exact floor/ball motion |

The star layer persists through several earlier units. Its simulation and depth
rules must be established from the original routine instead of selecting an
arbitrary generic distribution. Fifteen-second images do not establish complete
object lists or every transition. The final screen continues through the end of
the supplied movie; the original runtime determines its continuation behavior.
