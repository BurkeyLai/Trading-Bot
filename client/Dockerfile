#FROM node:14.18.1-alpine3.14
FROM node:14.18.1-alpine

#RUN apk add g++ make py3-pip

WORKDIR /app

COPY package.json yarn.lock ./

RUN apk --no-cache --virtual build-dependencies add \
    python2 \
    make \
    g++ \
    bash \
    && yarn

#RUN yarn
#RUN yarn cache clean && yarn upgrade && yarn

RUN yarn add sass
#RUN yarn add node-sass@4.14.1
#RUN yarn add node-sass

RUN apk del build-dependencies

COPY . .



CMD [ "yarn", "start" ]