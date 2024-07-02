FROM node:alpine

WORKDIR /home/container
ADD . /home/container
RUN apk add gcc imagemagick libwebp-tools ffmpeg go neofetch util-linux-misc

RUN npm i -g pm2
RUN npm install
#RUN go build -o bot
CMD pm2 start 'go run .' --watch '.go' --watch '/.go' --watch '//.go' && pm2 log
