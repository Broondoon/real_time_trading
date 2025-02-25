// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/views/market/fifty_fifty_stock_change_view.dart';
import 'package:flutter_frontend/views/market/market_controller.dart';
import 'package:provider/provider.dart';

class StockWidget extends StatelessWidget {
  const StockWidget({
    super.key,
    required this.stockId,
    required this.stockName,
    required this.stockPrice,
  });

  final String stockId;
  final String stockName;
  final String stockPrice;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () {
        print("$stockName pressed");
        Provider.of<MarketStateProvider>(
          context,
          listen: false,
        ).setStockShown(
          stockId,
          stockName,
          stockPrice,
        );
      },
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Column(
              children: [
                Text(
                  stockName,
                  style: MyAppStyle.regularFont,
                ),
                Text(
                  // ex: (GOOG)
                  '(${stockName.substring(4).toUpperCase()})',
                  style: MyAppStyle.regularFontLightGrey,
                ),
              ],
            ),
            Column(
              children: [
                Text(
                  // '\$293.70',
                  '\$$stockPrice',
                  style: MyAppStyle.regularFont,
                ),
                FiftyFiftyWidget(),
              ],
            ),
          ],
        ),
      ),
    );
  }
}