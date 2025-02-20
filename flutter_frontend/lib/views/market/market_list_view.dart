// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/market/stock_widget_view.dart';
import 'package:go_router/go_router.dart';

class MarketList extends StatefulWidget {
  const MarketList({
    super.key,
  });

  @override
  State<MarketList> createState() => _MarketListState();
}

class _MarketListState extends State<MarketList> {
  late List<Widget> _stockList;

  @override
  void initState() {
    super.initState();
    _stockList = [];
    _loadStockList();
  }

  Future<void> _loadStockList() async {
    // request from backend the stocks available
    //var stocks = getAvailStocks();
    List<String> stocks = ["1", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2"];
    for (String id in stocks) {
      _stockList.add(
        StockWidget(
          stockId: '999',
          stockName: 'Google',
          stockPrice: '293.70',
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 250,
      child: Drawer(
        child: Column(
          children: [
            InkWell(
              onTap: () {
                context.goNamed(homeRouteName);
              },
              child: SizedBox(
                height: 80,
                width: 250,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Row(
                    // mainAxisAlignment: MainAxisAlignment.,
                    children: [
                      Icon(
                        Icons.circle,
                      ),
                      SizedBox(
                        width: 8,
                      ),
                      Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            "Bob Duncan",
                            style: MyAppStyle.regularFont,
                          ),
                          Text(
                            "\$9999.99",
                            style: MyAppStyle.regularFont,
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
            Divider(
              height: 0,
              indent: 0,
            ),
            Expanded(
              // TODO: Make this list load one by one, rather than all at once, for bonus style points
              child: ListView.builder(
                itemCount: _stockList.length,
                itemBuilder: (context, index) {
                  return Column(
                    children: [
                      _stockList[index],
                      Divider(
                        indent: 2,
                        color: Colors.grey,
                        height: 0,
                      ),
                    ],
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}