// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_frontend/api_service.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:flutter_frontend/views/home/history/history_item_view.dart';
import 'package:provider/provider.dart';

class HistoryCard extends StatelessWidget {
  const HistoryCard({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    final APIService apiService = APIService(
      Provider.of<AuthController>(context),
    );

    List<Widget> historyItemList = [];

    // This is a little experimental, as I'm not too sure what I should grab
    //    to represent transaction history.
    Future<List<Response>> getTransactions() async {
      List<Response> walletAndStockTransactions = [];

      walletAndStockTransactions.add(
        await apiService.getWalletTransactions(),
      );
      walletAndStockTransactions.add(
        await apiService.getStockTransactions(),
      );

      return Future<List<Response>>.value(walletAndStockTransactions);
    }

    void populateHistoryItemList(List<Map> data) {
      List<Map> walletTs = data[0]['data'];
      List<Map> stockTs = data[1]['data'];

      if (stockTs.length != walletTs.length) {
        print('>> Transaction mismatch between getWalletTs and getStockTs');
        return;
      }

      for (int i = 0; i < walletTs.length; i++) {        
        historyItemList.add(
          HistoryItem(
            totalPrice: walletTs[i]['amount'],
            stockPrice: stockTs[i]['stock_price'],
            quantity: stockTs[i]['quantity'],
            timestamp: stockTs[i]['time_stamp'],
          ),
        );
      }
    }

    return Expanded(
      child: Card(
        child: FutureBuilder<void>(
          future: getTransactions(),
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
              final List<Response> responses = snapshot.data as List<Response>;
              final List<Map> data = [responses[0].data, responses[1].data];

              if (
                data[0] is Map && data[0].containsKey('success') && data[0]['success'] == true
                && data[1] is Map && data[1].containsKey('success') && data[1]['success'] == true
              ) {
                populateHistoryItemList(data);
                
                return Column(
                  children: [
                    Text(
                      'History',
                      style: MyAppStyle.largeFont,
                    ),
                    Expanded(
                      child: ListView.builder(
                        itemCount: historyItemList.length,
                        itemBuilder: (context, index) {
                          return historyItemList[index];
                        }
                      )
                    ),
                  ],
                );
              }
              else {
                print(">> Unexpected response behaviour.");
                return Column(
                  children: [
                    Text(
                      'History',
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