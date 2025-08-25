create table
  if not exists dgv_contracts_abi (
    id varchar(50) not null,
    chain_id int not null,
    address varchar(255) not null,
    type varchar(30) not null,
    implementation varchar(255),
    abi text,
    ctime timestamp default now (),
    utime timestamp,
    primary key (id)
  );
