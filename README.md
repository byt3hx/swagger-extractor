# Description

Swagger extractor is a tool which will extracts API endpoints from Swagger specification , sends requests to your proxy server via a simple web interface, making API testing more efficient.

# Features:
- Simple Web UI
- Support both openapi 2.0 & 3.0
- Custom Header
- Parse the request to proxy server(burp,caido,zap)
- Console Output
- Support All HTTP method, except DELETE

# Installation
```
git clone https://github.com/byt3hx/swagger-extractor.git
```
# Usuage
Run the following command and navigate to http://localhost:8081
```
go run extract.go
```
![](https://github.com/byt3hx/swagger-extractor/blob/b8b8c2a323a67efe7dc0a0204e182a509750d65e/Screenshot%202566-12-20%20at%2019.55.59.png)

# Example
![](https://github.com/byt3hx/swagger-extractor/blob/b9687f12e9e205510a466151486d0893e8e9577b/Screenshot%202566-12-20%20at%2019.53.58.png)

# Note
Please note that this tool is intended for local machine use only and is not ready for production deployment. I am not liable for any misuse or potential damage resulting from its use.
