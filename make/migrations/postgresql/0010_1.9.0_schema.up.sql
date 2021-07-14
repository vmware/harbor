/* add table for CVE whitelist */
CREATE TABLE cve_whitelist
(
  id            SERIAL PRIMARY KEY NOT NULL,
  project_id    int,
  creation_time timestamp default CURRENT_TIMESTAMP,
  update_time   timestamp default CURRENT_TIMESTAMP,
  expires_at    bigint,
  items         text               NOT NULL,
  UNIQUE (project_id)
);

CREATE TABLE blob
(
  id            SERIAL PRIMARY KEY NOT NULL,
  /*
     digest of config, layer, manifest
  */
  digest        varchar(255)        NOT NULL,
  content_type  varchar(1024)       NOT NULL,
  size          bigint              NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  UNIQUE (digest)
);

/* add the table for project and blob */
CREATE TABLE project_blob (
 id SERIAL PRIMARY KEY NOT NULL,
 project_id int NOT NULL,
 blob_id int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_project_blob UNIQUE (project_id, blob_id)
);

CREATE TABLE artifact
(
  id            SERIAL PRIMARY KEY NOT NULL,
  project_id    int                NOT NULL,
  repo          varchar(255)       NOT NULL,
  tag           varchar(255)       NOT NULL,
  /*
     digest of manifest
  */
  digest        varchar(255)       NOT NULL,
  /*
     kind of artifact, image, chart, etc..
  */
  kind          varchar(255)       NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  push_time     timestamp,
  CONSTRAINT unique_artifact UNIQUE (project_id, repo, tag)
);

/* add the table for relation of artifact and blob */
CREATE TABLE artifact_blob
(
  id            SERIAL PRIMARY KEY NOT NULL,
  digest_af     varchar(255)       NOT NULL,
  digest_blob   varchar(255)       NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  CONSTRAINT unique_artifact_blob UNIQUE (digest_af, digest_blob)
);

/* add quota table */
CREATE TABLE quota
(
  id            SERIAL PRIMARY KEY NOT NULL,
  reference     VARCHAR(255)       NOT NULL,
  reference_id  VARCHAR(255)       NOT NULL,
  hard          JSONB              NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  update_time   timestamp default CURRENT_TIMESTAMP,
  UNIQUE (reference_id,reference)
);

/* add quota usage table */
CREATE TABLE quota_usage
(
  id            SERIAL PRIMARY KEY NOT NULL,
  reference     VARCHAR(255)       NOT NULL,
  reference_id  VARCHAR(255)       NOT NULL,
  used          JSONB              NOT NULL,
  creation_time timestamp default CURRENT_TIMESTAMP,
  update_time   timestamp default CURRENT_TIMESTAMP,
  UNIQUE (reference_id,reference)
);

/* only set quota and usage for 'library', and let the sync quota handling others. */
INSERT INTO quota (reference, reference_id, hard, creation_time, update_time)
SELECT 'project',
       CAST(project_id AS VARCHAR),
       '{"count": -1, "storage": -1}',
       NOW(),
       NOW()
FROM project
WHERE name = 'library' and deleted = 'f';

INSERT INTO quota_usage (id, reference, reference_id, used, creation_time, update_time)
SELECT id,
       reference,
       reference_id,
       '{"count": 0, "storage": 0}',
       creation_time,
       update_time
FROM quota;

create table retention_policy
(
  id              serial PRIMARY KEY NOT NULL,
  scope_level     varchar(20),
  scope_reference integer,
  trigger_kind    varchar(20),
  data            text,
  create_time     time,
  update_time     time
);

create table retention_execution
(
  id         serial PRIMARY KEY NOT NULL,
  policy_id  integer,
  dry_run    boolean,
  trigger    varchar(20),
  start_time timestamp
);

create table retention_task
(
  id           SERIAL NOT NULL,
  execution_id integer,
  repository   varchar(255),
  job_id       varchar(64),
  status       varchar(32),
  status_code  integer,
  status_revision integer,
  start_time   timestamp default CURRENT_TIMESTAMP,
  end_time     timestamp default CURRENT_TIMESTAMP,
  total        integer,
  retained     integer,
  PRIMARY KEY (id)
);

create table schedule
(
  id            SERIAL NOT NULL,
  job_id        varchar(64),
  status        varchar(64),
  creation_time timestamp default CURRENT_TIMESTAMP,
  update_time   timestamp default CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

/*add notification policy table*/
create table notification_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 enabled boolean NOT NULL DEFAULT true,
 description text,
 targets text,
 event_types text,
 creator varchar(256),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 CONSTRAINT unique_project_id UNIQUE (project_id)
 );

/*add notification job table*/
 CREATE TABLE notification_job (
 id SERIAL NOT NULL,
 policy_id int NOT NULL,
 status varchar(32),
 /* event_type is the type of trigger event, eg. pushImage, pullImage, uploadChart... */
 event_type varchar(256),
 /* notify_type is the type to notify event to user, eg. HTTP, Email... */
 notify_type varchar(256),
 job_detail text,
 job_uuid varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

ALTER TABLE replication_task ADD COLUMN status_revision int DEFAULT 0;
DELETE FROM project_metadata WHERE deleted = TRUE;
ALTER TABLE project_metadata DROP COLUMN deleted;
