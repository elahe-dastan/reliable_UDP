[![Build Status](https://cloud.drone.io/api/badges/elahe-dastan/reliable_UDP/status.svg)](https://cloud.drone.io/elahe-dastan/reliable_UDP)

# Reliable UDP

This project uses udp sockets to transfer files but it uses acknowledgments to be reliable like TCP, no congestion <br/>
control has benn done.

## Approaches

I have tried to implement different approaches to do this <br/>
1-Stop and wait<br/>
2-Go back N<br/>
3-Selective repeat
