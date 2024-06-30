# FROM arm64v8/alpine:3.8
FROM --platform=linux/amd64 alpine:3.8
###############################################################################
#                                INSTALLATION
###############################################################################
RUN apk add --no-cache tzdata
ENV WORKDIR                 /app
ADD resource                $WORKDIR/resource
# ADD ./temp/linux_arm64/main $WORKDIR/main
ADD ./temp/linux_amd64/main $WORKDIR/main
RUN chmod +x $WORKDIR/main

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
CMD ./main
