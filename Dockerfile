FROM golang:latest
WORKDIR /home/container
ADD buildings .
ADD templates templates
HEALTHCHECK --interval=1m --timeout=3s \
  CMD curl -f http://localhost:8080/ || exit 1
EXPOSE 8080
CMD [ "./buildings", "serve" ]
