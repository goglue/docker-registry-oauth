FROM scratch

EXPOSE 4444

ADD docker-registry-oauth /docker-registry-oauth

ENTRYPOINT ["/docker-registry-oauth"]
