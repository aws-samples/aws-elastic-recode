FROM alpine
MAINTAINER SuWei <suwei007@gmail.com>

RUN set -xe \
    && apk add --no-cache ca-certificates 
                          
RUN apk add --no-cache bash

WORKDIR /app

COPY ./controlplane  /app/controlplane 

COPY ./docker-entriypoint.sh  /docker-entriypoint.sh

COPY ./home.html /app/home.html

COPY ./static  /app/static 

RUN chmod +x /docker-entriypoint.sh

EXPOSE 80

CMD ["/docker-entriypoint.sh"]