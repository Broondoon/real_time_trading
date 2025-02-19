// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class AccountCard extends StatelessWidget {
  const AccountCard({
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
                'Account',
                style: MyAppStyle.largeFont
              ),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.circle,
                  ),
                  Text(
                    'Bob Duncan',
                    style: MyAppStyle.regularFont,
                  ),
                ],
              ),
              Padding(
                padding: EdgeInsets.all(16.0),
                child: Card(
                  child: Padding(
                    padding: EdgeInsets.all(8.0),
                    child: SingleChildScrollView(
                      child: Text(
                        'This is my decription, my bio. This describes who I am. For I am a person who trades stock with vigour and finesse. You won\'t find a better trader out there. Believe it!',
                        style: MyAppStyle.regularFont,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          )
        ),
      ),
    );
  }
}