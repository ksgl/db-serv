FROM ubuntu:18.04

# docker build -t db .
# docker run -p 5000:5000 --name db -t db

ENV PGSQLVER 10
ENV DEBIAN_FRONTEND 'noninteractive'

RUN echo 'Europe/Moscow' > '/etc/timezone'

RUN apt-get -y update
RUN apt install -y gcc git wget
RUN apt install -y postgresql-$PGSQLVER

RUN wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -xvf go1.11.2.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /forum-server
COPY . .

EXPOSE 5000

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --echo-all --command "CREATE USER ksu WITH SUPERUSER PASSWORD 'pswd';" &&\
    createdb -O ksu parkdb &&\
    psql --dbname=parkdb --echo-all --command 'CREATE EXTENSION IF NOT EXISTS citext;' &&\
    psql parkdb -f /forum-server/database/scripts/scheme.psql &&\
    /etc/init.d/postgresql stop


RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGSQLVER/main/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "fsync = off" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "synchronous_commit = off" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "shared_buffers = 512MB" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "random_page_cost = 1.0" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "wal_level = minimal" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "max_wal_senders = 0" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf

EXPOSE 5432

USER root

RUN go build /forum-server/cmd/main.go
CMD service postgresql start && ./main