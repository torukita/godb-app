# README

aws RDS(postgres/mysql)のマルチAZ環境のfail over時に切り替わり状況を見るための簡易アプリ

##  準備

mysqlおよびpostgresでsampledbという名前のdatabaseを利用する。事前に作成しておく。

```
create database sampledb;
```

## 設定ファイル

DBへの接続情報をyaml形式で記載する。下記は例

postgres driverではfailover時に問題があった。(queryがcontextのtimeoutでcanncelせず、stackしたままになる)  
そこで、[pq-timeouts](https://github.com/Kount/pq-timeouts)というpostgres driverのwrapper driverを使う。  
これを使うことでDSNでreate_timeoutを指定することが可能になり、errorを検知できるようになる。  

それぞれで試してみるとよろし。  

 - db1はmysql driverを使用
 - db2はpostgres driverを使用
 - db3はpq-timeouts drvierを使用

```
# --- dbconf.yml ---
db1:
    driver: mysql
    open: root:root@tcp(hostx.xxx.amazon.com:3306)/sampledb?parseTime=true&&timeout=5s

db2:
    driver: postgres
    open: host=hosty.xxx.amazon.com dbname=sampledb user=root password=root port=5432 sslmode=disable connect_timeout=5

db3:
    driver: pq-timeouts
    open: host=hosty.xxx.amazon.com dbname=sampledb user=root password=root port=5432 sslmode=disable connect_timeout=5 read_timeout=2000 write_timeout=4000
```

 - mysql
     tableにtimestampを使用しているので、parseTimeオプションを指定
     failover時にDBへの再接続をする際にtimeoutするよう5sec指定(dns名が切り替わると接続可能になる)
 
 - postgres/pq-timeouts
     connect_timeoutはmysqlのtimeoutと同じ。defaultだとdns名が切り替わっているのに、syn/ackを待ち続けてしまう。
     read_timeoutおよびwrite_timeoutはcontext timeoutが効かない問題を回避 (pq-timeouts driverのときのみ有効なオプション)

## 使用

### テーブル作成

sampledb上にsample_dataというテーブルを作成する

```
% godb-app -c <path/dbconf.yml> -e db3 prepare
```

### テーブル削除

```
% godb-app -c <path/dbconf.yml> -e db3 cleanup
```
databaseは消さないので、注意

### テーブル内のデータ削除

```
% godb-app -c <path/dbconf.yml> -e db3 reset
```

### 実行

```
% godb-app -c <path/dbconf.yml> -e db3 run -n 100 -i 1s
```

### データ表示

```
% godb-app -c <path/dbconf.yml> -e db3 dump
```

### failover試す

```
% godb-app -c <path/dbconf.yml> -e db3 run -n 100 -i 1s --reconnect
```

1秒ごとにdataをinsertしていく。100回継続。--reconnectオプションはfailoverしてみるときにつけましょう。  
実行中にRDSをfailoverまたは再起動等する。errorが表示され、切り替わり正常に1回書き込みできると、終了する。  

## memo

プログラム上では、INSERT時に、context timeoutを設定(2sec)している。

## 例

```
$ ../bin/godb-app-linux --log-level debug -e db3 run -n 100 -i 1s --reconnect
DEBU[0000] Added data (1)
DEBU[0001] Added data (2)
DEBU[0002] Added data (3)
DEBU[0003] Added data (4)
DEBU[0004] Added data (5)
DEBU[0005] Added data (6)
DEBU[0006] Added data (7)
DEBU[0007] Added data (8)
DEBU[0008] Added data (9)
DEBU[0009] Added data (10)
DEBU[0010] Added data (11)
DEBU[0011] Added data (12) <-- このへんでfailover開始
DEBU[0012] Added data (13)
DEBU[0013] Added data (14)
DEBU[0014] Added data (15)
DEBU[0015] Added data (16)
ERRO[0023] [Aug  2 12:07:10.135] Failed data (17). it started at [Aug  2 12:07:03.134] AddMemo: read tcp 10.5.1.175:36362->10.5.16.111:5432: i/o timeout      <--- read_timeoutでerrorを検知(7secくらい)
ERRO[0028] [Aug  2 12:07:15.136] Connect: dial tcp 10.5.16.111:5432: i/o timeout  <-- connect_timeoutでerror (dnsが変わっていない)
ERRO[0033] [Aug  2 12:07:20.136] Connect: dial tcp 10.5.16.111:5432: i/o timeout  <-- connect_timeoutでerror (dnsが変わっていない)
ERRO[0038] [Aug  2 12:07:25.136] Connect: dial tcp 10.5.16.111:5432: i/o timeout  <-- 5secでerrorになってる
ERRO[0043] [Aug  2 12:07:30.137] Connect: dial tcp 10.5.16.111:5432: i/o timeout
ERRO[0048] [Aug  2 12:07:35.137] Connect: dial tcp 10.5.16.111:5432: i/o timeout
WARN[0048] [Aug  2 12:07:35.162] db connection recovered again by reconnect=6 times <-- dnsが更新されconnect成功 (切り替わり25secくらい)
INFO[0048] [Aug  2 12:07:35.173] Added data (17) after reconnection  (再接続後書き込めたので終了)
```

mysqlの場合

```
DEBU[0019] Added data (20)
DEBU[0020] Added data (21)
DEBU[0021] Added data (22)
DEBU[0022] Added data (23)
DEBU[0023] Added data (24)
ERRO[0026] [Aug  2 12:22:02.090] Failed data (25). it started at [Aug  2 12:22:00.090] AddMemo: context deadline exceeded <-- context timeoutが2secで有効になっている
ERRO[0031] [Aug  2 12:22:07.091] Connect: dial tcp 10.5.15.54:3306: i/o timeout  <-- DSNでのtimeout指定のため
ERRO[0036] [Aug  2 12:22:12.091] Connect: dial tcp 10.5.15.54:3306: i/o timeout
ERRO[0041] [Aug  2 12:22:17.092] Connect: dial tcp 10.5.15.54:3306: i/o timeout
ERRO[0046] [Aug  2 12:22:22.092] Connect: dial tcp 10.5.15.54:3306: i/o timeout
ERRO[0051] [Aug  2 12:22:27.093] Connect: dial tcp 10.5.15.54:3306: i/o timeout
ERRO[0056] [Aug  2 12:22:32.093] Connect: dial tcp 10.5.15.54:3306: i/o timeout
WARN[0056] [Aug  2 12:22:32.110] db connection recovered again by reconnect=7 times  <-- 30secくらいで切り替わり
INFO[0056] [Aug  2 12:22:32.116] Added data (25) after reconnection
```

