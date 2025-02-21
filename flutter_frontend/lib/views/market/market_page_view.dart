// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/views/market/content_view.dart';
import 'package:flutter_frontend/views/market/market_controller.dart';
import 'package:flutter_frontend/views/market/market_list_view.dart';
import 'package:provider/provider.dart';

class MarketPage extends StatelessWidget {
  const MarketPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      // appBar: AppBar(
      //   title: const Text('Market'),
      // ),
      body: ChangeNotifierProvider<MarketStateProvider> (
        create: (context) => MarketStateProvider(),
        child: Row(
          children: [
            MarketList(),
            StockContent(),
          ],
        ),
      ),
    );
  }
}