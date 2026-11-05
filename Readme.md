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
Uses a standard go lang server with graceful shutdown. Use --tls=True flag to use https. Defaults to http. Create certificates for TLS connectios in the webServer/certs folder
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=localhost"
Capabilities are
1. Check inclusion of a complex number. /check-inclusion
curl -v "https://localhost:8080/check-inclusion?X=1.25&Y=0.87"
2. Generate an image and sever throught HTTPS
