// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/home/portfolio/port_stock_item_view.dart';
import 'package:go_router/go_router.dart';

class PortfolioCard extends StatefulWidget {
  const PortfolioCard({
    super.key,
  });

  @override
  State<PortfolioCard> createState() => _PortfolioCardState();
}

class _PortfolioCardState extends State<PortfolioCard> {
  @override
  Widget build(BuildContext context) {

    // TODO: replace this with a function call to API
    List<String> listData = [
      'One',
      'Two',
      'Three',
      'Foiur',
    ];

    // TODO: the future return type may need to be 
    late Future<void> portfolioFuture;

    @override
    void initState() {
      super.initState();
      portfolioFuture;
    }

    return Expanded(
      child: Card(
        child: FutureBuilder<void>(
          future: portfolioFuture,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return Center(
                child: const CircularProgressIndicator(),
              );
            }
            else if (snapshot.hasError) {
              print('>>> Connection error: ${snapshot.error}');
              return Center(
                child: const Text(
                  'Something has gone terribly wrong - connection error.',
                  style: MyAppStyle.largeFont,
                ),
              );
            }
            else if (snapshot.connectionState == ConnectionState.done) {
              return Column(
                children: [
                  Text(
                    'Portfolio',
                    style: MyAppStyle.largeFont
                  ),
                  Expanded(
                    child: ListView.builder(
                      itemCount: listData.length,
                      itemBuilder: (context, index) {
                        return PortfolioItem(itemData: listData[index]);
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