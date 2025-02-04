package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPi(t *testing.T) {
	currencyUSD := util.USD
	currencyEUR := util.EUR
	fromAccountUSD := accountWithCurrency(currencyUSD)
	toAccountUSD := accountWithCurrency(currencyUSD)
	fromAccountEUR := accountWithCurrency(currencyEUR)
	toAccountEUR := accountWithCurrency(currencyEUR)

	amount := util.RandomMoney()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(1).Return(fromAccountUSD, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountUSD.ID)).Times(1).Return(toAccountUSD, nil)
				arg := db.TransferTxParams{
					FromAccountID: fromAccountUSD.ID,
					ToAccountID:   toAccountUSD.ID,
					Amount:        amount,
				}

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},

		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": int64(1001),
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(1001))).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountUSD.ID)).Times(0)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   int64(1001),
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(1).Return(fromAccountUSD, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(1001))).Times(1).Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccountEUR.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountEUR.ID)).Times(1).Return(fromAccountEUR, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountUSD.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountEUR.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(1).Return(fromAccountUSD, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountEUR.ID)).Times(1).Return(toAccountEUR, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        "XYZ",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountEUR.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          -amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountEUR.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountUSD.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": fromAccountUSD.ID,
				"to_account_id":   toAccountUSD.ID,
				"amount":          amount,
				"currency":        currencyUSD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountUSD.ID)).Times(1).Return(fromAccountUSD, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountUSD.ID)).Times(1).Return(toAccountUSD, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/createTransfer")
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}
}

func accountWithCurrency(currency string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: currency,
	}
}
