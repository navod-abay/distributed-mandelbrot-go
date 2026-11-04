This is an intentionally overcomplicated learning project built around generating an image of the Mandelbrot Set. No AI coding tools were used in development.

Available modes
Command line tool - generate an image in your local machine
Distributed computation - generate over multiple computing 
Web Server (WIP)


Distributed Calcualtion
This part is actually oversimplified and might have the opposite effects of taking more time to compute than running on a single node. I haven't taken two major variables into account when deleagting work to clients. They are network latency and single processor performance of a node. 


Distributed Benchmarking Prequisite
GCP secret manager api enabled.

### Web Server
Uses a standard go lang server with graceful shutdown. Capabilities are
1. Check inclusion of a complex number. /check-inclusion
2. Generate an image and sever throught HTTPS