// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_frontend/api_service.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:flutter_frontend/views/market/stock_widget_view.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

class MarketList extends StatefulWidget {
  const MarketList({
    super.key,
  });

  @override
  State<MarketList> createState() => _MarketListState();
}

class _MarketListState extends State<MarketList> {
  // @override
  // void initState() {
  //   super.initState();
  //   _stockList = [];
  //   _loadStockList();
  // }

  // Future<void> _loadStockList() async {
  //   // request from backend the stocks available
  //   //var stocks = getAvailStocks();
  //   List<String> stocks = ["1", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2"];
  //   for (String id in stocks) {
  //     _stockList.add(
  //       StockWidget(
  //         stockId: '999',
  //         stockName: 'Google',
  //         stockPrice: '293.70',
  //       ),
  //     );
  //   }
  // }

  @override
  Widget build(BuildContext context) {
    final APIService apiService = APIService(
      Provider.of<AuthController>(context, listen: false),
    );

    late List<Widget> stockList = [];

    void populateStockList(Map data) {
      for (Map stock in data['data']) {
        stockList.add(
          StockWidget(
            stockId: stock['stock_id'].toString(),
            stockName: stock['stock_name'],
            stockPrice: stock['current_price'].toString(),
          ),
        );
      }
    }

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
                          FutureBuilder(
                            future: apiService.getWalletBalance(), 
                            builder: (context, snapshot) {
                              if (snapshot.connectionState == ConnectionState.waiting) {
                                return Center(
                                  child: const CircularProgressIndicator(),
                                );
                              }
                              else if (snapshot.hasError) {
                                return Text(
                                  'Something has gone terribly wrong - ${snapshot.error}.',
                                  style: MyAppStyle.regularFontLightGrey,
                                );
                              }
                              else if (snapshot.connectionState == ConnectionState.done) {
                                final Response response = snapshot.data as Response;
                                final data = response.data;

                                // TODO: This doesn't update when the user's wallet changes! Eh, that's probably fine.
                                if (data is Map && data.containsKey('success') && data['success'] == true) {
                                  return Text(
                                    "\$${data['data'][0]['balance']}",
                                    style: MyAppStyle.regularFont,
                                  );
                                }
                                else {
                                  print(">> Unexpected response behaviour.");
                                  return Text(
                                    'Unexpected network error.',
                                    style: MyAppStyle.regularFontLightGrey,
                                  );
                                }
                              }
                              else {
                                return Text(
                                  'Something has gone terribly wrong - unhandled connection state.',
                                  style: MyAppStyle.regularFontLightGrey,
                                );
                              }
                            },
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
              child: FutureBuilder(
                future: apiService.getStockPrices(),
                builder: (context, snapshot) {
                  if (snapshot.connectionState == ConnectionState.waiting) {
                    return Center(
                      child: const CircularProgressIndicator(),
                    );
                  }
                  else if (snapshot.hasError) {
                    return Text(
                      'Something has gone terribly wrong - ${snapshot.error}.',
                      style: MyAppStyle.largeFont,
                    );
                  }
                  else if (snapshot.connectionState == ConnectionState.done) {
                    final Response response = snapshot.data as Response;
                    final data = response.data;

                    if (data is Map && data.containsKey('success') && data['success'] == true) {
                      populateStockList(data);

                      return ListView.builder(
                        itemCount: stockList.length,
                        itemBuilder: (context, index) {
                          return Column(
                            children: [
                              stockList[index],
                              Divider(
                                indent: 2,
                                color: Colors.grey,
                                height: 0,
                              ),
                            ],
                          );
                        },
                      );
                    }
                    else {
                      print(">> Unexpected response behaviour.");
                      return Text(
                        'Unexpected network error.',
                        style: MyAppStyle.regularFontLightGrey,
                      );
                    }
                  }
                  else {
                    return Text(
                      'Something has gone terribly wrong - unhandled connection state.',
                      style: MyAppStyle.largeFont,
                    );
                  }
                }
              ),
            ),
          ],
        ),
      ),
    );
  }
}