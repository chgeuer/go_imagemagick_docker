FROM alpine:3.9 as builder-base

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin && \
    apk --update add bash vim curl imagemagick imagemagick-dev git go pkgconf make gcc libc-dev && \
    cd /usr/include/ImageMagick-7 && \
    ln -s MagickWand wand && \
    ln -s MagickCore magick && \
    curl -sL https://glide.sh/get | sh && \
    go get gopkg.in/gographics/imagick.v3/imagick && \
    rm -rf /var/cache/apk/*

#####################################

FROM builder-base as builder

#VOLUME ["/app"]

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR $GOPATH

COPY run.go .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ./app . && \
   ls -als .

CMD ["magick", "-help"]

###############################

FROM alpine:3.9

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR $GOPATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin && \
    apk --update add bash imagemagick && \
    rm -rf /var/cache/apk/*

COPY --from=builder /go/app .
# COPY IMG_20190131_065124.jpg /go/input.jpg

CMD ["app"]
