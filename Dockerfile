FROM scratch
COPY config /config
ADD reaper /
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/
CMD ["/reaper"]
EXPOSE 30003


