# mysql-sheets
Allows Google Sheets access from a MySQL client

**Work in progress**

To use with the example doc https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms

1. Download `credentials.json` from https://developers.google.com/sheets/api/quickstart/go#step_1_turn_on_the

2. Start the MySheetSQL server:
```
go run cmd/mysheetsql/main.go serve --id 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms
```

3. Connect with any MySQL client, user `user`, password `pass`:
```
$ mysql -u user -h 127.0.0.1 -P 3306 -ppass

mysql> SHOW DATABASES;
+----------------------------------------------+
| Database                                     |
+----------------------------------------------+
| 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms |
+----------------------------------------------+
1 row in set (0.00 sec)

mysql> SHOW TABLES;
+------------+
| Table      |
+------------+
| Class Data |
+------------+
1 row in set (0.63 sec)

mysql> SELECT * FROM `Class Data`;
+--------------+--------+--------------+------------+---------+--------------------------+
| Student Name | Gender | Class Level  | Home State | Major   | Extracurricular Activity |
+--------------+--------+--------------+------------+---------+--------------------------+
| Alexandra    | Female | 4. Senior    | CA         | English | Drama Club               |
| Andrew       | Male   | 1. Freshman  | SD         | Math    | Lacrosse                 |
| Anna         | Female | 1. Freshman  | NC         | English | Basketball               |
| Becky        | Female | 2. Sophomore | SD         | Art     | Baseball                 |
| Benjamin     | Male   | 4. Senior    | WI         | English | Basketball               |
| Carl         | Male   | 3. Junior    | MD         | Art     | Debate                   |
| Carrie       | Female | 3. Junior    | NE         | English | Track & Field            |
| Dorothy      | Female | 4. Senior    | MD         | Math    | Lacrosse                 |
| Dylan        | Male   | 1. Freshman  | MA         | Math    | Baseball                 |
| Edward       | Male   | 3. Junior    | FL         | English | Drama Club               |
| Ellen        | Female | 1. Freshman  | WI         | Physics | Drama Club               |
| Fiona        | Female | 1. Freshman  | MA         | Art     | Debate                   |
| John         | Male   | 3. Junior    | CA         | Physics | Basketball               |
| Jonathan     | Male   | 2. Sophomore | SC         | Math    | Debate                   |
| Joseph       | Male   | 1. Freshman  | AK         | English | Drama Club               |
| Josephine    | Female | 1. Freshman  | NY         | Math    | Debate                   |
| Karen        | Female | 2. Sophomore | NH         | English | Basketball               |
| Kevin        | Male   | 2. Sophomore | NE         | Physics | Drama Club               |
| Lisa         | Female | 3. Junior    | SC         | Art     | Lacrosse                 |
| Mary         | Female | 2. Sophomore | AK         | Physics | Track & Field            |
| Maureen      | Female | 1. Freshman  | CA         | Physics | Basketball               |
| Nick         | Male   | 4. Senior    | NY         | Art     | Baseball                 |
| Olivia       | Female | 4. Senior    | NC         | Physics | Track & Field            |
| Pamela       | Female | 3. Junior    | RI         | Math    | Baseball                 |
| Patrick      | Male   | 1. Freshman  | NY         | Art     | Lacrosse                 |
| Robert       | Male   | 1. Freshman  | CA         | English | Track & Field            |
| Sean         | Male   | 1. Freshman  | NH         | Physics | Track & Field            |
| Stacy        | Female | 1. Freshman  | NY         | Math    | Baseball                 |
| Thomas       | Male   | 2. Sophomore | RI         | Art     | Lacrosse                 |
| Will         | Male   | 4. Senior    | FL         | Math    | Debate                   |
+--------------+--------+--------------+------------+---------+--------------------------+
30 rows in set (1.24 sec)

mysql> SELECT Gender, Major, COUNT(*) AS num FROM `Class Data` GROUP BY Gender, Major ORDER BY Major,Gender;
+--------+---------+------+
| Gender | Major   | num  |
+--------+---------+------+
| Female | Art     |    3 |
| Male   | Art     |    4 |
| Female | English |    4 |
| Male   | English |    4 |
| Female | Math    |    4 |
| Male   | Math    |    4 |
| Female | Physics |    4 |
| Male   | Physics |    3 |
+--------+---------+------+
8 rows in set (1.04 sec)
```