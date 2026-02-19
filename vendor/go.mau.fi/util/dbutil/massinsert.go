// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"fmt"
	"regexp"
	"strings"
)

// Array is an interface for small fixed-size arrays.
// It exists because generics can't specify array sizes: https://github.com/golang/go/issues/44253
type Array interface {
	[0]any | [1]any | [2]any | [3]any | [4]any | [5]any | [6]any | [7]any | [8]any | [9]any |
		[10]any | [11]any | [12]any | [13]any | [14]any | [15]any | [16]any | [17]any | [18]any | [19]any |
		[20]any | [21]any | [22]any | [23]any | [24]any | [25]any | [26]any | [27]any | [28]any | [29]any
}

// MassInsertable represents a struct that contains dynamic values for a mass insert query.
type MassInsertable[T Array] interface {
	GetMassInsertValues() T
}

// MassInsertBuilder contains pre-validated templates for building mass insert SQL queries.
type MassInsertBuilder[Item MassInsertable[DynamicParams], StaticParams Array, DynamicParams Array] struct {
	queryTemplate       string
	placeholderTemplate string
}

// NewMassInsertBuilder creates a new MassInsertBuilder that can build mass insert database queries.
//
// Parameters in mass insert queries are split into two types: static parameters
// and dynamic parameters. Static parameters are the same for all items being
// inserted, while dynamic parameters are different for each item.
//
// The given query should be a normal INSERT query for a single row. It can also
// have ON CONFLICT clauses, as long as the clause uses `excluded` instead of
// positional parameters.
//
// The placeholder template is used to replace the `VALUES` part of the given
// query. It should contain a positional placeholder ($1, $2, ...) for each
// static placeholder, and a fmt directive (`$%d`) for each dynamic placeholder.
//
// The given query and placeholder template are validated here and the function
// will panic if they're invalid (e.g. if the `VALUES` part of the insert query
// can't be found, or if the placeholder template doesn't have the right things).
// The idea is to use this function to populate a global variable with the mass
// insert builder, so the panic will happen at startup if the query or
// placeholder template are invalid (instead of returning an error when trying
// to use the query later).
//
// Example:
//
//	type Message struct {
//		ChatID    int
//		RemoteID  string
//		MXID      id.EventID
//		Timestamp time.Time
//	}
//
//	func (msg *Message) GetMassInsertValues() [3]any {
//		return [3]any{msg.RemoteID, msg.MXID, msg.Timestamp.UnixMilli()}
//	}
//
//	const insertMessageQuery = `INSERT INTO message (chat_id, remote_id, mxid, timestamp) VALUES ($1, $2, $3, $4)`
//	var massInsertMessageBuilder = dbutil.NewMassInsertBuilder[Message, [2]any](insertMessageQuery, "($1, $%d, $%d, $%d, $%d)")
//
//	func DoMassInsert(ctx context.Context, messages []*Message) error {
//		query, params := massInsertMessageBuilder.Build([1]any{messages[0].ChatID}, messages)
//		return db.Exec(ctx, query, params...)
//	}
func NewMassInsertBuilder[Item MassInsertable[DynamicParams], StaticParams Array, DynamicParams Array](
	singleInsertQuery, placeholderTemplate string,
) *MassInsertBuilder[Item, StaticParams, DynamicParams] {
	var dyn DynamicParams
	var stat StaticParams
	totalParams := len(dyn) + len(stat)
	mainQueryVariablePlaceholderParts := make([]string, totalParams)
	for i := 0; i < totalParams; i++ {
		mainQueryVariablePlaceholderParts[i] = fmt.Sprintf(`\$%d`, i+1)
	}
	mainQueryVariablePlaceholderRegex := regexp.MustCompile(fmt.Sprintf(`\(\s*%s\s*\)`, strings.Join(mainQueryVariablePlaceholderParts, `\s*,\s*`)))
	queryPlaceholders := mainQueryVariablePlaceholderRegex.FindAllString(singleInsertQuery, -1)
	if len(queryPlaceholders) == 0 {
		panic(fmt.Errorf("invalid insert query: placeholders not found"))
	} else if len(queryPlaceholders) > 1 {
		panic(fmt.Errorf("invalid insert query: multiple placeholders found"))
	}
	for i := 0; i < len(stat); i++ {
		if !strings.Contains(placeholderTemplate, fmt.Sprintf("$%d", i+1)) {
			panic(fmt.Errorf("invalid placeholder template: static placeholder $%d not found", i+1))
		}
	}
	if strings.Contains(placeholderTemplate, fmt.Sprintf("$%d", len(stat)+1)) {
		panic(fmt.Errorf("invalid placeholder template: non-static placeholder $%d found", len(stat)+1))
	}
	fmtParams := make([]any, len(dyn))
	for i := 0; i < len(dyn); i++ {
		fmtParams[i] = fmt.Sprintf("$%d", len(stat)+i+1)
	}
	formattedPlaceholder := fmt.Sprintf(placeholderTemplate, fmtParams...)
	if strings.Contains(formattedPlaceholder, "!(EXTRA string=") {
		panic(fmt.Errorf("invalid placeholder template: extra string found"))
	}
	for i := 0; i < len(dyn); i++ {
		if !strings.Contains(formattedPlaceholder, fmt.Sprintf("$%d", len(stat)+i+1)) {
			panic(fmt.Errorf("invalid placeholder template: dynamic placeholder $%d not found", len(stat)+i+1))
		}
	}
	return &MassInsertBuilder[Item, StaticParams, DynamicParams]{
		queryTemplate:       strings.Replace(singleInsertQuery, queryPlaceholders[0], "%s", 1),
		placeholderTemplate: placeholderTemplate,
	}
}

// Build constructs a ready-to-use mass insert SQL query using the prepared templates in this builder.
//
// This method always only produces one query. If there are lots of items,
// chunking them beforehand may be required to avoid query parameter limits.
// For example, SQLite (3.32+) has a limit of 32766 parameters by default,
// while Postgres allows up to 65535. To find out if there are too many items,
// divide the maximum number of parameters by the number of dynamic columns in
// your data and subtract the number of static columns.
//
// Example of chunking input data:
//
//	var mib dbutil.MassInsertBuilder
//	var db *dbutil.Database
//	func MassInsert(ctx context.Context, ..., data []T) error {
//		return db.DoTxn(ctx, nil, func(ctx context.Context) error {
//			for _, chunk := range exslices.Chunk(data, 100) {
//				query, params := mib.Build(staticParams)
//				_, err := db.Exec(ctx, query, params...)
//				if err != nil {
//					return err
//				}
//			}
//			return nil
//		}
//	}
func (mib *MassInsertBuilder[Item, StaticParams, DynamicParams]) Build(static StaticParams, data []Item) (query string, params []any) {
	var itemValues DynamicParams
	params = make([]any, len(static)+len(itemValues)*len(data))
	placeholders := make([]string, len(data))
	for i := 0; i < len(static); i++ {
		params[i] = static[i]
	}
	fmtParams := make([]any, len(itemValues))
	for i, item := range data {
		baseIndex := len(static) + len(itemValues)*i
		itemValues = item.GetMassInsertValues()
		for j := 0; j < len(itemValues); j++ {
			params[baseIndex+j] = itemValues[j]
			fmtParams[j] = baseIndex + j + 1
		}
		placeholders[i] = fmt.Sprintf(mib.placeholderTemplate, fmtParams...)
	}
	query = fmt.Sprintf(mib.queryTemplate, strings.Join(placeholders, ", "))
	return
}
