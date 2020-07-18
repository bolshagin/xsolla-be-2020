CREATE TABLE `Sessions` (
    `SessionID` INT NOT NULL AUTO_INCREMENT,
    `SessionToken` NVARCHAR(4000) NOT NULL,
    `Amount` FLOAT NOT NULL,
    `Purpose` NVARCHAR(4000) NULL,
    `CreatedAt` DATETIME NOT NULL,
    `ClosedAt` DATETIME NULL,
    PRIMARY KEY (`SessionID`)
);
