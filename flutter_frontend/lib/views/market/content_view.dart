// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_frontend/api_service.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:flutter_frontend/views/market/market_controller.dart';
import 'package:provider/provider.dart';

class StockContent extends StatefulWidget {
  const StockContent({
    super.key,
  });

  @override
  State<StockContent> createState() => _StockContentState();
}

class _StockContentState extends State<StockContent> {
  final _buyAmountController = TextEditingController();
  final _sellAmountController = TextEditingController();
  final _sellPriceController = TextEditingController();

  // @override
  // void initState() {
  //   super.initState();
  // }

  // TODO: properly dispose of editing controllers
  @override
  void dispose() {
    _buyAmountController.dispose();
    _sellAmountController.dispose();
    _sellPriceController.dispose();
    super.dispose();
  }

  // Initially I did this with amount passed in.
  // But to mirror wallet_card_view, I'm not.
  void purchaseStock(APIService apiService, String stockId, String stockName, String stockPrice) async {
    if (_buyAmountController.text.isEmpty) {
      return;
    }

    int amount = int.parse(_buyAmountController.text);

    Response response = await apiService.placeStockOrder(
      int.parse(stockId),
      true, // isBuy
      'MARKET', // orderType
      amount, // quantity
      int.parse(stockPrice),
    );
    final data = response.data;

    if (data is Map && data.containsKey('success') && data['success'] == true) {
      print("Success on ordering $amount stock!");
      // TODO: should have a snackbar popup
    }
  }

  void sellStock(APIService apiService, String stockId, String stockName, String stockPrice) async {
    if (_sellAmountController.text.isEmpty || _sellPriceController.text.isEmpty) {
      return;
    }

    int amount = int.parse(_sellAmountController.text);
    int listingPrice = int.parse(_sellPriceController.text);

    Response response = await apiService.placeStockOrder(
      int.parse(stockId),
      false, // isBuy
      'LIMIT', // orderType
      amount, // quantity
      listingPrice, // price
    );
    final data = response.data;

    if (data is Map && data.containsKey('success') && data['success'] == true) {
      print("Success on ordering $amount stock!");
      // TODO: should have a snackbar popup
    }
  }

  Widget showContent(String stockId, String stockName, String stockPrice) {
    final APIService apiService = APIService(
      Provider.of<AuthController>(context, listen: false),
    );

    if (stockId != '-1') {
      return Column(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Text(
                stockName,
                style: MyAppStyle.titleFont
              ),
            ),
          ),
          Text(
            stockPrice,
            style: MyAppStyle.largeFont,
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Column(
                children: [
                  ElevatedButton(
                    onPressed: () {
                      print("Pressed purchase.");
                      purchaseStock(
                        apiService,
                        stockId,
                        stockName,
                        stockPrice,
                      );
                    },
                    child: Text('Buy'),
                  ),
                  SizedBox(
                    height: 50,
                    width: 60,
                    child: Card(
                      child: Padding(
                        padding: const EdgeInsets.all(8.0),
                        child: TextField(
                          controller: _buyAmountController,
                          maxLength: 2,
                          decoration: const InputDecoration(
                            // labelText: 'Funds',
                            // labelStyle: MyAppStyle.regularFont,
                            hintText: 'x99',
                            hintStyle: MyAppStyle.regularFontLightGrey,
                            border: InputBorder.none,
                            counterText: '',
                          ),
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              Column(
                children: [
                  ElevatedButton(
                    onPressed: () {
                      print("Pressed sell.");
                      sellStock(
                        apiService,
                        stockId,
                        stockName,
                        stockPrice,
                      );
                    },
                    child: Text('Sell'),
                  ),
                  Row(
                    children: [
                      SizedBox(
                        height: 50,
                        width: 60,
                        child: Card(
                          child: Padding(
                            padding: const EdgeInsets.all(8.0),
                            child: TextField(
                              controller: _sellAmountController,
                              maxLength: 2,
                              decoration: const InputDecoration(
                                hintText: 'x99',
                                hintStyle: MyAppStyle.regularFontLightGrey,
                                border: InputBorder.none,
                                counterText: '',
                              ),
                              keyboardType: TextInputType.number,
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                              ],
                            ),
                          ),
                        ),
                      ),
                      SizedBox(
                        height: 50,
                        width: 80,
                        child: Card(
                          child: Padding(
                            padding: const EdgeInsets.all(8.0),
                            child: TextField(
                              controller: _sellPriceController,
                              maxLength: 2,
                              decoration: const InputDecoration(
                                hintText: '\$999.99',
                                hintStyle: MyAppStyle.regularFontLightGrey,
                                border: InputBorder.none,
                                counterText: '',
                              ),
                              keyboardType: TextInputType.number,
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                              ],
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              )
            ],
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Card(
                child: Center(
                  child: Text('Chart goes here'),
                ),
              ),
            ),
          ),
        ],
      );
    }
    else {
      return SizedBox();
    }
  }

  // final TextEditingController _buyAmountController;
  @override
  Widget build(BuildContext context) {
    // These will be listening for changes made to the selected stock in the Market State provider
    final String _stockId = Provider.of<MarketStateProvider>(context).getCurrStockId;
    final String _stockName = Provider.of<MarketStateProvider>(context).getCurrStockName;
    final String _stockPrice = Provider.of<MarketStateProvider>(context).getCurrStockPrice;

    return Expanded(
      child: showContent(_stockId, _stockName, _stockPrice),
    );
  }
}
