// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/main.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:go_router/go_router.dart';

class MarketPage extends StatefulWidget {
  const MarketPage({super.key});

  @override
  State<MarketPage> createState() => _MarketPageState();
}

class _MarketPageState extends State<MarketPage> {

  late List<Widget> _stockList;

  @override
  void initState() {
    super.initState();
    _stockList = [];
    _loadStockList();
  }

  // TODO: properly implement
  Future<void> _loadStockList() async {
    // request from backend the stocks available
    //var stocks = getAvailStocks();
    List<String> stocks = ["1", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2"];
    for (String id in stocks) {
      _stockList.add(
        StockWidget(
          // TODO: add params to StockWidget
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      // appBar: AppBar(
      //   title: const Text('Market'),
      // ),
      body: Row(
        children: [
          SizedBox(
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
          ),
          // Main page content area
          Center(
            child: Text('HOY', style: MyAppStyle.titleFont),
          )
        ],
      ),
    );
  }
}

class StockWidget extends StatelessWidget {
  const StockWidget({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () {
        print("Tapped a GOOG");
      },
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Column(
              children: [
                Text(
                  'Google',
                  style: MyAppStyle.regularFont,
                ),
                Text(
                  '(GOOG)',
                  style: MyAppStyle.regularFontLightGrey,
                ),
              ],
            ),
            Column(
              children: [
                Text(
                  '\$293.70',
                  style: MyAppStyle.regularFont,
                ),
                Row(
                  children: [
                    Icon(
                      Icons.keyboard_double_arrow_up,
                    ),
                    Text(
                      '12.2%',
                      style: MyAppStyle.regularFontLightGrey,
                    )
                  ],
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}