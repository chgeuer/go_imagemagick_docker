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
    rm -rf /var/cache/apk/*

#####################################

FROM builder-base as builder

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
ENV WD $GOPATH/src/github.com/chgeuer/go_imagemagick_docker/

WORKDIR $WD

COPY glide.yaml glide.lock ./
RUN  glide install

COPY run.go ./
RUN  CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ./app .

# ###############################

FROM alpine:3.9
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
ENV WD $GOPATH/src/github.com/chgeuer/go_imagemagick_docker/

WORKDIR $WD

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin && \
    apk --update add bash imagemagick && \
    rm -rf /var/cache/apk/*

COPY --from=builder $WD/app ./

# COPY IMG_20190131_065124.jpg ./input.jpg
CMD ["app"]
