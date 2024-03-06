# Stream Test

1. Starts receiver on localhost:9999.
2. Sends test file with bitrate to local receiver.

`bin/streamey -url localhost:9999 -filepath test_files/test_tone.mp3 -bitrate 192000 -test`

# Stream

Sends test file with bitrate to url.

`bin/streamey -url example.com:9999 -filepath test_files/test_tone.mp3 -bitrate 192000`
