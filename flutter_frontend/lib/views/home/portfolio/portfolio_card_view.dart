// TODO: remove these ignores for when this class is finished
// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class PortfolioCard extends StatelessWidget {
  const PortfolioCard({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Center(
          child: Column(
            children: [
              Text(
                'Portfolio',
                style: MyAppStyle.largeFont
              ),
              Expanded(
                child: Card(),
              ),
              ElevatedButton(
                onPressed: () => {},
                child: Text(
                  'Search the Market',
                  style: MyAppStyle.regularFont,
                )
              ),
              SizedBox(
                height: 8.0,
              )
            ],
          )
        ),
      ),
    );
  }
}
