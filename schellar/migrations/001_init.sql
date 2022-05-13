create table schedule(
  schedule_name varchar(100) primary key,
  is_enabled boolean not null,
  workflow_status varchar(20) not null,
  workflow_name varchar(100) not null,
  workflow_version varchar(20) not null,
  workflow_context json not null,
  cron_string varchar(10) not null,
  parallel_runs boolean not null,
  check_warning_seconds int not null,
  from_date timestamptz,
  to_date timestamptz,
  correlation_id varchar(100),
  task_to_domain json,
  last_update timestamptz not null
);

---- create above / drop below ----

--drop table schedule;
