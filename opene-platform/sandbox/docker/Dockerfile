# syntax=docker/dockerfile:1.7
# OpenE sandbox: egen, reproducerbar Open ePlatform-distribution för test.
# Byggs från låsta källor (build/versions.env). Ingen hemlighet bakas in i imagen.

ARG JDK_IMAGE=eclipse-temurin:21-jdk-jammy
ARG JRE_IMAGE=tomcat:9.0.123-jre21-temurin-jammy

# ------------------------------------------------------------------ bygg
FROM ${JDK_IMAGE} AS build
RUN apt-get update \
 && apt-get install -y --no-install-recommends subversion git curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /opt/build
COPY build/ ./
# Källhämtningen cachas i ett eget lager så att ombyggen går snabbt.
RUN --mount=type=cache,target=/root/.gradle \
    curl --fail --location --retry 3 --connect-timeout 30 \
      https://github.com/frebergfabian/OpenE-AI-skill/releases/download/v0.5.0-dev/svn-r3190-r1649-r15.tar.gz \
      -o /opt/build/svn-mirror.tar.gz \
 && OPENE_SVN_MIRROR=/opt/build/svn-mirror.tar.gz \
    OPENE_SVN_MIRROR_SHA256=744d32447e3d39b91db333110074c004bcde6981401ea42c233caeed866c63ad \
    bash build-opene.sh /opt/work /opt/dist

# ------------------------------------------------------------------ runtime
FROM ${JRE_IMAGE} AS runtime
RUN apt-get update \
 && apt-get install -y --no-install-recommends mariadb-client openssl curl \
 && rm -rf /var/lib/apt/lists/* \
 && rm -rf /usr/local/tomcat/webapps/* \
 && groupadd --system --gid 10001 opene \
 && useradd --system --uid 10001 --gid opene --home-dir /var/lib/opene opene \
 && mkdir -p /var/lib/opene/files /var/lib/opene/tmp /var/log/opene \
 && chown -R opene:opene /var/lib/opene /var/log/opene /usr/local/tomcat/logs /usr/local/tomcat/work /usr/local/tomcat/temp

COPY --from=build --chown=opene:opene /opt/dist/webapp /usr/local/tomcat/webapps/ROOT
COPY --from=build /opt/dist/sql /opt/opene/sql
COPY --from=build /opt/dist/BUILDINFO /opt/opene/BUILDINFO
COPY docker/sql/ /opt/opene/sql/
COPY docker/config.xml.tmpl docker/entrypoint.sh docker/seed-users.sh /opt/opene/
RUN chmod 0755 /opt/opene/*.sh \
 && chown opene:opene /usr/local/tomcat/webapps/ROOT/WEB-INF

USER opene
ENV CATALINA_OPTS="-Dfile.encoding=ISO-8859-1 -Djava.awt.headless=true -Xms256m -Xmx1g" \
    OPENE_FILESTORE=/var/lib/opene/files
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=180s --retries=20 \
  CMD curl -fsS -o /dev/null http://127.0.0.1:8080/ || exit 1
ENTRYPOINT ["/opt/opene/entrypoint.sh"]
CMD ["catalina.sh", "run"]
