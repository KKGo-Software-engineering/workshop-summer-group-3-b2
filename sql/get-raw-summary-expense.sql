SELECT
    date_trunc('day', date)::date AS transaction_date,
    SUM(amount) AS total_amount,
    COUNT(*) AS record_count
FROM
    "transaction"
WHERE
    transaction_type = 'expense' AND spender_id = 1
GROUP BY
    date_trunc('day', date)::date
ORDER BY
    transaction_date;
