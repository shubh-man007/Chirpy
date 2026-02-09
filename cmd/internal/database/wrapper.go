package database

import (
	"context"
	"database/sql"
)

func (q *Queries) AcceptFriendRequestSafely(ctx context.Context, arg AcceptFriendRequestParams) error {
	res, err := q.db.ExecContext(ctx, acceptFriendRequest, arg.UserID, arg.FriendID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (q *Queries) RejectFriendRequestSafely(
	ctx context.Context,
	arg RejectFriendRequestParams,
) error {
	res, err := q.db.ExecContext(
		ctx,
		rejectFriendRequest,
		arg.UserID,
		arg.FriendID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
