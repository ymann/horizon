FROM desertbit/golang-gb


ADD . /project
RUN cd /project
RUN gb build

CMD /project/bin/horizon

