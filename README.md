# SCION DVB-T Streaming POC

This project aims to read data from a DVB-T receiver Raspberry Pi Hat and to stream the data into SCION Network.

Currently, we simulate an incoming signal using cvlc until a raspberry pi is ready. 

## Setup
At first, setup dependencies and build sender.go and receiver.go. Then setup the SCION VM and create private key and certificate using `openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes`

Put some mp4 video file in `./sample.mp4` (will be removed later).

 Then start at first the simulation, then the sender.go and at last the receiver.go.

`cd /vagrant`
simulation: `cvlc sample.mp4 --sout-keep --sout "#transcode{vcodec=x264,vb=800,scale=1,acodec=mp4a,ab=128,channels=2}:duplicate{dst=std{access=http,mux=ts,dst=localhost:8008}}"`
sender: `./sender --cert cert.pem --key key.pem`
receiver: `./receiver`