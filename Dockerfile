FROM alpine:3.9

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin && \
    apk --update add bash vim imagemagick imagemagick-dev git go pkgconf make gcc libc-dev && \
    cd /usr/include/ImageMagick-7 && \
    ln -s MagickWand wand && \
    ln -s MagickCore magick && \
    go get gopkg.in/gographics/imagick.v3/imagick && \
    rm -rf /var/cache/apk/*

#VOLUME ["/app"]

WORKDIR $GOPATH

CMD ["magick", "-help"]
