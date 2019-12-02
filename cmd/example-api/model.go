package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //go-lint-ignore
	"log"
)

const colNames string = "ID, CustomerID, DueDate, Amount, Description"

type invoicesModel struct {
	db *sql.DB
}

func newInvoicesModel(db *sql.DB) invoicesModel {
	return invoicesModel{db: db}
}

func (model *invoicesModel) create(ctx context.Context, i invoice) (invoice, error) {
	result, err := model.db.ExecContext(ctx,
		"INSERT INTO invoices (CustomerID, DueDate, Amount, Description) VALUES (?, ?, ?, ?)",
		i.CustomerID,
		i.DueDate,
		i.Amount,
		i.Description)

	if err != nil {
		return invoice{}, err
	}
	ID, err := result.LastInsertId()
	if err != nil {
		return invoice{}, err
	}

	fmt.Println(fmt.Sprintf("ID=%v", ID))

	return model.getByID(ctx, int(ID))
}

func parseRow(scanFn func(...interface{}) error) (invoice, error) {
	var i invoice = invoice{}
	var description sql.NullString
	var dueDate sql.NullTime

	if err := scanFn(
		&i.ID,
		&i.CustomerID,
		&dueDate,
		&i.Amount,
		&description); err != nil {
		return invoice{}, err
	}

	if dueDate.Valid {
		i.DueDate = dueDate.Time
	}

	if description.Valid {
		i.Description = description.String
	}
	return i, nil
}

func (model *invoicesModel) getAll(ctx context.Context) ([]invoice, error) {
	rows, err := model.db.QueryContext(ctx, fmt.Sprintf("SELECT %v FROM invoices", colNames))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	invoices := []invoice{}
	for rows.Next() {
		i, err := parseRow(rows.Scan)
		if err != nil {
			return []invoice{}, err
		}
		invoices = append(invoices, i)
	}
	if err := rows.Err(); err != nil {
		return []invoice{}, err
	}

	return invoices, nil
}

func (model *invoicesModel) getByID(ctx context.Context, ID int) (invoice, error) {
	row := model.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %v FROM invoices WHERE ID=?", colNames), ID)
	i, err := parseRow(row.Scan)

	switch {
	case err == sql.ErrNoRows:
		return invoice{}, NotFoundError(fmt.Sprintf("Invoice with ID=%d not found", ID))
	case err != nil:
		return invoice{}, err
	default:
		return i, nil
	}
}
