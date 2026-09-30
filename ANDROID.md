# Mental Hangover on Android

The mobile host runs the same Go renderer, embedded assets and DCK v1.0.4
music facade as the desktop command. Its clock remains at the original 50 Hz.
The fixed 352 by 272 scene is centered with its aspect ratio intact. Android
provides landscape orientation, immersive display, screen-awake behavior and
pause/resume handling; production logic remains in Go.

## Build and run

Existing requirements: Go, Java 17, Android SDK 36, NDK r28 and an authorized
ARM64 device. The script selects Ebitengine/ebitenmobile v2.9.11, AGP 8.10.1
and the committed Gradle 8.11.1 wrapper. It does not install analysis tools.

```sh
./scripts/run-android.sh --build-only
./scripts/run-android.sh
```

The debug APK is `android/app/build/outputs/apk/debug/app-debug.apk`.
Generated AARs, APKs, build/cache directories and machine-specific settings are
excluded locally. The application ID is `com.olivierh.mentalhangover`.

The 30 September 2026 ARM64 build embeds DCK v1.0.4 after the shared color and
contour migrations. Its four native ELF load segments and the APK library
placement pass 16 KiB alignment. Desktop checks cover 699 matching complete
frames with the published module and all 293 source mode/level combinations
over 4,096 RGB12 colors. The package was installed on the Pixel 10a on
30 September and traversed the whole director through the checkerboard finale
without an observed crash. Five-second measurements recorded 49.2–51.0 ticks/s,
58.1–60.1 displayed frames/s and a peak Go heap of 53.3 MiB. These heap figures
exclude native/GPU memory; runtime cadence is separate from visual fidelity.

## Device verification

A verification launch can show the demo over a locked screen without
dismissing the keyguard or changing security settings:

```sh
adb shell am start -S -W -n com.olivierh.mentalhangover/.MainActivity --ei mental_verify_tick 0
```

Tick zero retains normal music playback and logs scene transitions plus
TPS/FPS and memory every five seconds. A positive tick up to 30,000 seeks a
muted reconstruction before continuing at 50 Hz. This provides repeatable
checks of late scenes without changing normal playback.
Verification requests apply at startup. The `-S` option stops the existing
process before selecting a different checkpoint; `-W` waits for the launch.
Resuming an existing activity preserves its scene clock.

```sh
adb shell am start -S -W -n com.olivierh.mentalhangover/.MainActivity --ei mental_verify_tick 22000
adb logcat -s GoLog:I AndroidRuntime:E
```

Closing/relaunching the application starts a fresh production. Backgrounding
suspends both rendering and Ebitengine audio. No touch input is required by
this linear production. A locked-device runtime check does not establish
complete visual fidelity; the source/video comparisons are tracked separately.
