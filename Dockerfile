FROM ubuntu:18.04

# docker build -t fsdb .
# docker run -p 5000:5000 --name fsdb -t fsdb
# docker run -d --memory 1G --log-opt max-size=1M --log-opt max-file=3 --name fsdb -p 5000:5000 fsdb

# useful variables
ENV PGSQLVER 10
ENV DEBIAN_FRONTEND 'noninteractive'

RUN echo 'Europe/Moscow' > '/etc/timezone'

# aptitude update
RUN apt-get -y update
RUN apt install -y gcc git wget
RUN apt install -y postgresql-$PGSQLVER

# install golang
RUN wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -xvf go1.11.2.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

# retrieving the project
WORKDIR /forum-server
COPY . .

EXPOSE 5000

# starting postgres, creating user/db, adjusting configs
USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --echo-all --command "CREATE USER ksu WITH SUPERUSER PASSWORD 'pswd';" &&\
    createdb -O ksu parkdb &&\
    psql --dbname=parkdb --echo-all --command 'CREATE EXTENSION IF NOT EXISTS citext;' &&\
    psql parkdb -f /forum-server/database/scripts/scheme_refactored.psql &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGSQLVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf
RUN echo "shared_buffers = 512MB" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432

# starting
USER root

RUN go build /forum-server/cmd/main.go
CMD service postgresql start && ./main