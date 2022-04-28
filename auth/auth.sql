create schema auth;

set search_path = 'public';

CREATE TABLE account (
                                id character varying(36) NOT NULL unique,
                                email character varying(100),
                                first_name character varying(100),
                                is_enabled boolean,
                                last_name character varying(100),
                                login character varying(100),
                                password character varying(100),
                                role_name character varying(30)
);

ALTER TABLE account OWNER TO tramites;

create table token (
                              id             bigint                   not null,
                              value          text                     not null,
                              account_id     character varying(36)		not null,
                              creationtime   timestamp with time zone not null,
                              expirationtime timestamp with time zone not null,
                              lasttime       timestamp with time zone not null,
                              primary key (id),
                              foreign key (account_id) references public.account (id)
);
ALTER TABLE token OWNER TO tramites;


create index i_token_1
  on token (expirationtime);

create unique index u_token_1
  on token (value);

