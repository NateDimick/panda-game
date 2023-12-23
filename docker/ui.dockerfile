FROM node:20-alpine

COPY frontend/ /panda-game-workdir/

WORKDIR /panda-game-workdir

RUN npm run build

CMD [ "node", "build/index.js"]