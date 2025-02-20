// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_frontend/resources/app_style.dart';
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

  @override
  void initState() {
    super.initState();
  }

  // final TextEditingController _buyAmountController;
  @override
  Widget build(BuildContext context) {
    // These will be listening for changes made to the selected stock in the Market State provider
    final String _stockId = Provider.of<MarketStateProvider>(context).getCurrStockId;
    final String _stockName = Provider.of<MarketStateProvider>(context).getCurrStockName;
    final String _stockPrice = Provider.of<MarketStateProvider>(context).getCurrStockPrice;

    return Expanded(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Text(
                _stockName,
                style: MyAppStyle.titleFont
              ),
            ),
          ),
          Text(
            _stockPrice,
            style: MyAppStyle.largeFont,
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Column(
                children: [
                  ElevatedButton(
                    onPressed: () => {},
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
                    onPressed: () => {},
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
                              controller: _sellAmountController,
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
      ),
    );
  }
}
