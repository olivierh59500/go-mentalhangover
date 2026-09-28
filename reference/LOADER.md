# Original loader reconstruction

The supplied disk has 80 cylinders, two sides and eleven 512-byte sectors per
track. Its nominal DOS0 root is empty. The code in the bootblock copies 101
longwords from disk offset `0x30` to Amiga address `0x20000`, then calls the
inherited trackdisk request with byte offset `0x400`, length `0x8000` and
destination `0x30000`.

The program at disk offset `0x400` copies `0x1e59` longwords from `0x41e` to
memory `0x21e` and starts the relocated resident code there. The resident's
routine at `0x7ba` reads cylinders, processing both sides in each iteration.
Its MFM-decoded 32-bit data words are XORed with `0x21051972`. The Go extractor
applies that XOR to the sector bytes already decoded in the ADF.

The transfer order and destination addresses are retained in
[`assets/raw/transfers.json`](../assets/raw/transfers.json). The first three
loads are the eagle, music data and author/ribbon unit. Later code modules are
loaded at `0x9000`, except patterned vectors at `0x8000`. The cube uses
cylinder 39 after the patterned unit has loaded cylinders 41–43.

The module occupies exactly 143,136 bytes, including its 31 patterns and sample
bank, at the beginning of the music transfer. Those bytes are identical to the
supplied `madness.mod`. Remaining bytes in that transfer are retained
for subsequent asset identification.

## Graphics and clocks identified so far

- The eagle has four separated bitplanes, a 44-byte row stride, width 352 and
  height 168. Its plane offsets are `0`, `0x1ce0`, `0x39c0`, `0x56a0`. The
  resident's fifteen 12-bit palette words at `0x68e` supply its colors.
- The display copper uses horizontal bounds `0x71`–`0x1d1` and vertical bounds
  `0x22`–`0x132`. Bitmap rows advance between copper waits at `0x40` and `0xe8`.
- The serif font starts at resident address `0x19ce`. Its rows interleave three
  planes, each with a 360-byte storage row. Glyph cells occupy six bytes and
  23 visible rows; 59 supported characters are ASCII space through `Z`.
  Their individual advances come from the word table at `0x17de`, plus one.
  The seven target palette words start at `0x19c0`.
- The three hardware star banks start at resident address `0x9c6`, occupy
  `0x16c` bytes each and have 45 single-row sprite headers per bank. The video
  interrupt at `0x3fe` advances their horizontal header bytes by 3, 2 and 1
  every frame. DCK's fixed-step field preserves these byte-wrap clocks.
- The original module replay is called by the same video interrupt. The
  resident's wait routine at `0x1728` resets and compares a VBL counter at
  `0x7f800`. These clocks will provide exact scene durations; the initial video
  scene map remains provisional.

Ghidra's existing Motorola processor was used to follow entry points in the
relocated resident and all nine executable effect blocks. No newly installed
tool, emulator runtime or prerecorded video is required by the Go port.
