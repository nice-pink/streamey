# Read Stream

1. Read stream from url

`bin/readey -url https://fluxmusic.api.radiosphere.io/channels/90s/stream.mp3`

# Validate stream

## Encoding validation

Define encoding expectations in config.json!

`bin/readey -url https://fluxmusic.api.radiosphere.io/channels/90s/stream.mp3 -validate audio -config bin/config.json`

## Private bit validation - duration

This can be used for e.g. streaming system e2e testing. Set private bit in audio header of first frame and validate,
that the distance between the frames with the private bit set never changes. Thus, no frames are missing in output and
no frames are repeated.

`bin/readey -url https://fluxmusic.api.radiosphere.io/channels/90s/stream.mp3 -validate privateBit`
