FROM node:lts-alpine

WORKDIR /app

COPY . .

RUN yarn install --production

RUN apk add --update tzdata

RUN chmod +x /app/schedule.sh
RUN chmod +x /app/start.sh
RUN source /app/schedule.sh

CMD [ "crond", "-l", "2", "-f" ]