create schema audit;

set search_path = 'audit';

create sequence auditentryseq;



create table auditentry (
  id             bigint                   not null
    constraint auditentry_pkey
      primary key,
  entityname     varchar(62)              not null,
  recordid       bigint                   not null,
  createdtime    timestamp with time zone not null,
  createdby      varchar(255)             not null,
  createdcontext varchar(255),
  operation      varchar(1)               not null,
  updatedtime    timestamp with time zone not null,
  updatedby      varchar(255)             not null,
  updatedcontext varchar(255)
);



create index i_auditentry_2
  on auditentry (updatedtime);

create unique index u_auditentry_1
  on auditentry (entityname, recordid);

create function register_audit() returns trigger
  language plpgsql
as
$$
DECLARE
  inentityname         VARCHAR(255);
  DECLARE inoperation  VARCHAR(1);
  DECLARE inusername   VARCHAR;
  DECLARE auditentryid BIGINT;
  DECLARE inrecordid   BIGINT;
  DECLARE inactive     VARCHAR;
  DECLARE incontext    VARCHAR;
BEGIN
  inactive := current_setting('tkt.audit_off', TRUE);

  IF inactive IS NOT NULL AND inactive = 'true'
  THEN
    RETURN NEW;
  END IF;

  inentityname := TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;
  inoperation := substr(TG_OP, 1, 1);
  inusername := current_setting('tkt.user_name', TRUE);
  incontext := current_setting('tkt.context', TRUE);

  IF inoperation = 'I'
  THEN
    inrecordid := NEW.id;
  ELSE
    inrecordid := OLD.id;
  END IF;


  IF inusername IS NULL
  THEN
    SELECT current_user INTO inusername;
    inusername := '@' || inusername;
  END IF;

  SELECT "ae"."id" INTO auditentryid
  FROM audit.auditentry ae
  WHERE "ae"."entityname" = inentityname
    AND "ae"."recordid" = inrecordid;
  IF auditentryid IS NULL
  THEN
    INSERT INTO audit.auditentry (id, entityname, recordid, createdtime, createdby, createdcontext,
                                  operation, updatedtime, updatedby, updatedcontext)
    VALUES (nextval('audit.auditentryseq'), inentityname, inrecordid, now(),
            inusername, incontext, inoperation, now(), inusername, incontext);
  ELSE
    UPDATE audit.auditentry
    SET operation      = inoperation,
        updatedby      = inusername,
        updatedtime    = now(),
        updatedcontext = incontext
    WHERE id = auditentryid;
  END IF;
  RETURN NEW;
END;
$$;



