FROM elasticrecode/ffmpeg:cpu-v7
MAINTAINER Su Wei <suwei007@gmail.com>



WORKDIR /app

RUN  mkdir /video




COPY ./ffmpeg_job.sh /app/ffmpeg_job.sh
COPY ./vmaf_job.sh /app/vmaf_job.sh
COPY ./pre_stop.sh /app/pre_stop.sh
COPY ./sqs_consumer.py /app/sqs_consumer.py
COPY ./sqs_producer.py /app/sqs_producer.py
COPY ./common /app/common

RUN chmod +x /app/ffmpeg_job.sh
RUN chmod +x /app/vmaf_job.sh
RUN chmod +x /app/pre_stop.sh
RUN chmod +x /app/sqs_consumer.py

CMD ["/app/sqs_consumer.py"]
