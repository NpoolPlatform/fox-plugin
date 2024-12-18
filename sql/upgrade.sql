START TRANSACTION;

-- fil
SELECT
  @pid := `id`,
  @reserved_amount := `reserved_amount`,
  @pre_sale := `pre_sale`,
  @logo := `logo`,
  @for_pay := `for_pay`
FROM
  `fox_coininfo`.`coin_infos`
WHERE
  `name` = 'filecoin'
  AND `env` = 'test';

UPDATE
  `fox_coininfo`.`coin_infos`
SET
  `id` = UUID()
WHERE
  `name` = 'filecoin'
  AND `env` = 'test';

UPDATE
  `fox_coininfo`.`coin_infos`
SET
  `id` = @pid,
  `reserved_amount` = @reserved_amount,
  `pre_sale` = @pre_sale,
  `logo` = @logo,
  `for_pay` = @for_pay
WHERE
  `name` = 'tfilecoin'
  AND `env` = 'test';

-- btc
SELECT
  @pid := `id`,
  @reserved_amount := `reserved_amount`,
  @pre_sale := `pre_sale`,
  @logo := `logo`,
  @for_pay := `for_pay`
FROM
  `fox_coininfo`.`coin_infos`
WHERE
  `name` = 'bitcoin'
  AND `env` = 'test';

UPDATE
  `fox_coininfo`.`coin_infos`
SET
  `id` = UUID()
WHERE
  `name` = 'bitcoin'
  AND `env` = 'test';

UPDATE
  `fox_coininfo`.`coin_infos`
SET
  `id` = @pid,
  `reserved_amount` = @reserved_amount,
  `pre_sale` = @pre_sale,
  `logo` = @logo,
  `for_pay` = @for_pay
WHERE
  `name` = 'tbitcoin'
  AND `env` = 'test';

COMMIT;