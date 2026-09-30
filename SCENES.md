# Reference scene inventory

The supplied recording lasts 465.026 seconds at 50 frames/s. This initial map
used sparse observations; it now incorporates exact-frame samples and decoded
routine counts for every effect. Video extraction uses explicit
frame selection: an fps resampling filter can shift a contact-sheet timestamp
by half its sampling interval. The director keeps source operation counts;
disk-loading holds are calibrated to the recording, not filesystem speed.

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
| 288.76–355.58 s | Circular text over a raster-colored mountain silhouette | Owned contour scrolling mode, ByteWindow, TablePolar, retained image layers | Original font/wave/profile tables, command payloads, phase fades and mountain occlusion |
| 357.58–381.46 s | Contact text and projected blue spheres | Projected sprite batches, held image layers | Authored positions, size bank, steering and hardware text columns |
| 382.46–416.00 s | Perspective blue text and points | Owned contour scrolling mode, ByteWindow, RationalGrid, pixel batch | Original font and projection coefficients; separate point-plane OR controller remains a candidate |
| 418.96–430.72 s | Final reminder card and disk-loading interval | Bitmap pages, cue ranges | Original flash/fades, held reminder and last loader hold |
| 430.72 s onward | Scoopex logo, perspective checkerboard and bouncing balls | Retained mask and row-palette shader, projected sprite batches | Original palettes, logo and verified floor/ball motion |

The three hardware-star banks persist through the earlier units and return
around the final reminder. Later dual-playfield parts disable that sprite layer
and use their own verified point/sphere programs. The final screen continues
indefinitely, matching the original production's continuation behavior.
