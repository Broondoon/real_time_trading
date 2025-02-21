// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_frontend/api_service.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:flutter_frontend/views/home/portfolio/port_stock_item_view.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

class PortfolioCard extends StatelessWidget {
  const PortfolioCard({
    super.key,
  });

  @override
  Widget build(BuildContext context) {    
    // TODO: the future return type may need to be 
    // late Future<Response> portfolioFuture;
    final APIService apiService = APIService(
      Provider.of<AuthController>(context),
    );

    List<Widget> portfolioItemList = [];

    // void loadPortfolio() async {
    //   Response getResponse = await portfolioFuture();
    // }

    // @override
    // void initState() {
    //   super.initState();
    //   portfolioFuture = apiService.getPortfolio();
    //   loadPortfolio();
    // }

    void populatePortfolioList(Map data) {
      for (Map stock in data['data']) {
        portfolioItemList.add(
          PortfolioItem(
            stockName: stock['stock_name'],
            stockPrice: "\$999", //stock['???'].toString(), // TODO: turns out WE DON'T GET THIS! I should have read the requests better
            quantOwned: stock['quantity_owned'].toString(),
            isPending: false,
          ),
        );
      }
    }

    return Expanded(
      child: Card(
        child: FutureBuilder<void>(
          future: apiService.getPortfolio(),
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return Center(
                child: const CircularProgressIndicator(),
              );
            }
            else if (snapshot.hasError) {
              print('>> Connection error: ${snapshot.error}');
              return Center(
                child: const Text(
                  'Something has gone terribly wrong - connection error.',
                  style: MyAppStyle.largeFont,
                ),
              );
            }
            else if (snapshot.connectionState == ConnectionState.done) {
              final Response response = snapshot.data as Response;
              final data = response.data;

              if (data is Map && data.containsKey('success') && data['success'] == true) {
                populatePortfolioList(data);
                
                return Column(
                  children: [
                    Text(
                      'Portfolio',
                      style: MyAppStyle.largeFont,
                    ),
                    Expanded(
                      child: ListView.builder(
                        itemCount: portfolioItemList.length,
                        itemBuilder: (context, index) {
                          return portfolioItemList[index];
                        }
                      )
                    ),
                    ElevatedButton(
                      onPressed: () => context.goNamed(marketRouteName),
                      child: Text(
                        'Search the Market',
                        style: MyAppStyle.regularFont,
                      )
                    ),
                    SizedBox(
                      height: 8.0,
                    )
                  ],
                );
              }
              else {
                print(">> Unexpected response behaviour.");
                return Column(
                  children: [
                    Text(
                      'Portfolio',
                      style: MyAppStyle.largeFont,
                    ),
                    Text(
                      'Unexpected network error.',
                      style: MyAppStyle.regularFontLightGrey,
                    )
                  ],
                );
              }
            }
            else {
              return Center(
                child: const Text(
                  'Something has gone terribly wrong - Unhanddled connection state.',
                  style: MyAppStyle.largeFont,
                ),
              );
            }
          },
        ),
      ),
    );
  }
}