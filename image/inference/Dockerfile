FROM alpine:3.14 as BUILDER
RUN apk update && apk add --no-cache \
        git \
        bash \
        libc6-compat
WORKDIR /opt/build
COPY ./obsutil.tar.gz ./obsutil.tar.gz
RUN tar -xf ./obsutil.tar.gz

# real service
FROM python:3.9.13

WORKDIR /usr/src/app

COPY ./requirements.txt ./requirements.txt
COPY ./run.sh ./run.sh
COPY ./pretrain.py ./pretrain.py
COPY --from=BUILDER /opt/build/obsutil /usr/src/app/obsutil

RUN pip install --upgrade -i https://pypi.tuna.tsinghua.edu.cn/simple pip 
RUN pip install --user -i https://pypi.tuna.tsinghua.edu.cn/simple -r requirements.txt
RUN apt-get update
RUN apt-get install ffmpeg libsm6 libxext6 -y
RUN apt install libgl1-mesa-glx

EXPOSE 7860
