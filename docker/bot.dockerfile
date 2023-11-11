FROM python:alpine3.17

COPY bot/bot.py .

COPY bot/requirements.txt .

RUN pip install -r requirements.txt

ENTRYPOINT [ "python", "bot.py" ]
