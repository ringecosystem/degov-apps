create table
  if not exists dgv_user (
    id varchar(50) not null,
    address varchar(255) not null,
    email varchar(255),
    primary key (id)
  );

create unique index uq_dgv_user_address on dgv_user (address);

comment on table dgv_user is 'User table for degov apps';

comment on column dgv_user.address is 'wallet address';

create table
  if not exists dgv_dao (
    id varchar(50) not null,
    chain_id int not null,
    chain_name varchar(255) not null,
    name varchar(255) not null,
    code varchar(255) not null,
    seq int not null default 0,
    state varchar(50) not null, -- { ACTIVE, INACTIVE }
    tags text,
    config_link varchar(255) not null,
    time_syncd timestamp,
    ctime timestamp default now (),
    utime timestamp,
    primary key (id)
  );

create unique index uq_dgv_dao_code on dgv_dao (code);

comment on table dgv_dao is 'Available DAOs';

comment on column dgv_dao.chain_id is 'chain id';

comment on column dgv_dao.chain_name is 'chain name';

comment on column dgv_dao.name is 'DAO name';

comment on column dgv_dao.code is 'DAO code';

comment on column dgv_dao.tags is 'Optional tags for DAO categorization';

comment on column dgv_dao.config_link is 'DAO config link';

comment on column dgv_dao.time_syncd is 'last syncd time';

create table
  if not exists dgv_dao_config (
    id varchar(50) not null,
    code varchar(255) not null,
    config text not null,
    ctime timestamp default now (),
    utime timestamp,
    primary key (id)
  );

create unique index uq_dgv_dao_config_code on dgv_dao_config (code);

comment on table dgv_dao_config is 'DAO config table';

create table
  if not exists dgv_user_liked_dao (
    id varchar(50) not null,
    dao_code varchar(255) not null,
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    ctime timestamp default now (),
    primary key (id)
  );

create unique index uq_dgv_user_liked_dao_code_uid on dgv_user_liked_dao (dao_code, user_id);

create unique index uq_dgv_user_liked_dao_code_address on dgv_user_liked_dao (dao_code, user_address);

comment on table dgv_user_liked_dao is 'DAO like table';

comment on column dgv_user_liked_dao.dao_code is 'DAO code';

comment on column dgv_user_liked_dao.user_id is 'user id';

comment on column dgv_user_liked_dao.user_address is 'user address';

create table
  if not exists dgv_user_subscribed_dao (
    id varchar(50) not null,
    chain_id int not null,
    dao_code varchar(255) not null,
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    state varchar(50) not null, -- { SUBSCRIBED, UNSUBSCRIBED }
    enable_new_proposal int not null default 1,
    enable_voting_end_reminder int not null default 0,
    ctime timestamp default now (),
    primary key (id)
  );

create unique index uq_dgv_user_subscribe_uid_code on dgv_user_subscribed_dao (user_id, dao_code);

create unique index uq_dgv_user_subscribe_address_code on dgv_user_subscribed_dao (user_address, dao_code);

comment on table dgv_user_subscribed_dao is 'User subscribed DAO table';

comment on column dgv_user_subscribed_dao.user_id is 'user id';

comment on column dgv_user_subscribed_dao.user_address is 'user address';

comment on column dgv_user_subscribed_dao.dao_code is 'DAO code';

comment on column dgv_user_subscribed_dao.enable_new_proposal is 'enable new proposal notification';

comment on column dgv_user_subscribed_dao.enable_voting_end_reminder is 'enable voting end reminder';

create table
  if not exists dgv_user_subscribed_proposal (
    id varchar(50) not null,
    chain_id int not null,
    dao_code varchar(255) not null,
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    state varchar(50) not null, -- { ACTIVE, INACTIVE }
    proposal_id varchar(255) not null,
    ctime timestamp default now (),
    primary key (id)
  );

comment on table dgv_user_subscribed_proposal is 'User subscribed proposal table';

comment on column dgv_user_subscribed_proposal.user_address is 'user address';

create table
  if not exists dgv_notification_record (
    id varchar(50) not null,
    chain_id int not null,
    chain_name varchar(255) not null,
    dao_name varchar(255) not null,
    dao_code varchar(255) not null,
    type varchar(50) not null, -- { NEW_PROPOSAL, VOTE, STATUS, VOTE_END_REMINDER }
    target_id varchar(255), -- proposal id or vote id
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    status varchar(50) not null, -- { SENT_OK, SENT_FAIL }
    message text,
    retry_times int not null default 0, -- number of times to retry sending
    ctime timestamp default now (),
    primary key (id)
  );

comment on table dgv_notification_record is 'Notification record table';

comment on column dgv_notification_record.chain_id is 'chain id';

comment on column dgv_notification_record.chain_name is 'chain name';

comment on column dgv_notification_record.dao_name is 'DAO name';

comment on column dgv_notification_record.dao_code is 'DAO code';

comment on column dgv_notification_record.type is 'notification type';

comment on column dgv_notification_record.target_id is 'target id (proposal id or vote id)';

comment on column dgv_notification_record.user_id is 'user id';

comment on column dgv_notification_record.status is 'notification status';

create table
  if not exists dgv_user_channel (
    id varchar(50) not null,
    user_id varchar(50) not null,
    user_address varchar(255) not null,
    verified int not null default 0, -- whether the channel is verified
    channel_type varchar(50) not null, -- { EMAIL, SMS, PUSH }
    channel_value varchar(500) not null, -- email address or phone number
    payload text, -- additional data for the channel
    ctime timestamp default now (),
    primary key (id)
  );

comment on table dgv_user_channel is 'Notification channel settings for users';

comment on column dgv_user_channel.user_id is 'user id';

comment on column dgv_user_channel.channel_type is 'notification channel type';

comment on column dgv_user_channel.channel_value is 'notification channel value (email or phone)';

comment on column dgv_user_channel.payload is 'additional data for the channel';
