CREATE TABLE `adinfo` (
  `aid` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(45) NOT NULL,
  `createAt` timestamp(6) NOT NULL,
  `startAt` timestamp(6) NOT NULL,
  `endAt` timestamp(6) NOT NULL,
  `ageStart` int(10) NOT NULL,
  `ageEnd` int(10) NOT NULL,
  `gender` varchar(45) NOT NULL,
  `country` varchar(45) NOT NULL,
  `platform` varchar(45) NOT NULL,
  PRIMARY KEY (`aid`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci